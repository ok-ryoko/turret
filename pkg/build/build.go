// Copyright 2023 OK Ryoko
// SPDX-License-Identifier: Apache-2.0

package build

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ok-ryoko/turret/pkg/container"
	"github.com/ok-ryoko/turret/pkg/linux"
	"github.com/ok-ryoko/turret/pkg/linux/find"
	"github.com/ok-ryoko/turret/pkg/linux/user"
	"github.com/ok-ryoko/turret/pkg/spec"

	"github.com/containers/buildah"
	is "github.com/containers/image/v5/storage"
	"github.com/containers/storage"
	"github.com/containers/storage/pkg/archive"
	"github.com/sirupsen/logrus"
)

const manifestType string = "application/vnd.oci.image.manifest.v1+json"

// Execute runs the build pipeline.
func Execute(ctx context.Context, s spec.Spec, logger *logrus.Logger, options ExecuteOptions) error {
	storeOptions, err := storage.DefaultStoreOptionsAutoDetectUID()
	if err != nil {
		storeOptions = storage.StoreOptions{}
	}
	store, err := storage.GetStore(storeOptions)
	if err != nil {
		return fmt.Errorf("creating store: %w", err)
	}
	defer func() {
		layers, errShutdown := store.Shutdown(false)
		if errShutdown != nil {
			logger.Warnln("failed releasing driver resources")
			logger.Infoln(
				"the following layers may still be mounted:",
				strings.Join(layers, ", "),
			)
		}
	}()

	if refThis := s.This.Reference(); store.Exists(refThis) && !options.Force {
		return fmt.Errorf("image %s already exists", refThis)
	}

	buildahOptions := buildah.BuilderOptions{
		Capabilities: []string{},
		FromImage:    s.From.Reference(),
		Isolation:    buildah.IsolationOCIRootless,
		PullPolicy:   buildah.PullNever,
	}
	if options.LogCommands {
		buildahOptions.Logger = logger
	}
	if options.Pull {
		buildahOptions.PullPolicy = buildah.PullIfMissing
	}

	buildahBuilder, err := buildah.NewBuilder(ctx, store, buildahOptions)
	if err != nil {
		return fmt.Errorf("creating Buildah builder: %w", err)
	}
	logger.Debugf("created working container from image %s", buildahOptions.FromImage)

	ctr := container.Container{
		Builder: buildahBuilder,
		Logger:  logger,
	}
	defer func() {
		if !options.Keep {
			if removeErr := ctr.Remove(); removeErr != nil {
				logger.Warnln("failed deleting working container")
				logger.Infoln("please remove the container manually: buildah rm", ctr.ContainerID())
			}
		}
	}()
	logger.Debugf("created %s Linux working container", s.From.Distro)

	if ctr.Builder.OS() != "linux" {
		return fmt.Errorf("expected 'linux' image, got '%s' image", ctr.Builder.OS())
	}

	ctr.CommonOptions.LogCommands = options.LogCommands
	if s.From.Distro.Distro == linux.Debian {
		ctr.CommonOptions.Env = append(ctr.CommonOptions.Env, "DEBIAN_FRONTEND=noninteractive")
	}

	pckgFrontend, err := container.NewPackageFrontend(s.Backends.Package.Backend)
	if err != nil {
		return fmt.Errorf("creating package management interface: %w", err)
	}

	userFrontend, err := container.NewUserFrontend(s.Backends.User.Backend)
	if err != nil {
		return fmt.Errorf("creating user management interface: %w", err)
	}

	findCmdFactory, err := find.NewCommandFactory(s.Backends.Find.Backend)
	if err != nil {
		return fmt.Errorf("creating find command factory: %w", err)
	}

	if s.Packages.Upgrade {
		logger.Debugln("upgrading packages in the working container...")
		if err := upgradePackages(&ctr, pckgFrontend); err != nil {
			return fmt.Errorf("upgrading packages: %w", err)
		}
		logger.Debugln("upgrade command ran successfully")
	}

	if len(s.Packages.Install) > 0 {
		logger.Debugln("installing packages to the working container...")
		if err := installPackages(&ctr, pckgFrontend, s.Packages.Install); err != nil {
			return fmt.Errorf("installing packages: %w", err)
		}
		logger.Debugln("install command ran successfully")
	}

	if s.Packages.Clean {
		if err := cleanPackageCaches(&ctr, pckgFrontend); err != nil {
			return fmt.Errorf("cleaning package caches: %w", err)
		}
		logger.Debugln("clean command ran successfully")
	}

	if s.User != nil {
		createUserOptions := user.Options{
			ID:         s.User.ID,
			UserGroup:  s.User.UserGroup,
			Groups:     s.User.Groups,
			Comment:    s.User.Comment,
			CreateHome: s.User.CreateHome,
		}
		if err := createUser(&ctr, userFrontend, s.User.Name, createUserOptions); err != nil {
			return fmt.Errorf("creating nonroot user: %w", err)
		}
		logger.Debugf("created nonroot user")
	}

	if len(s.Copy) > 0 {
		for _, cp := range s.Copy {
			copyFilesOptions := copyFilesOptions{
				excludes:      cp.Excludes,
				mode:          cp.Mode,
				owner:         cp.Owner,
				removeSpecial: cp.RemoveS,
			}
			if err := copyFiles(&ctr, cp.Base, cp.Destination, cp.Sources, copyFilesOptions); err != nil {
				return fmt.Errorf("copying files: %w", err)
			}
		}
		logger.Debugln("file copy command(s) ran successfully")
	}

	if s.Security.SpecialFiles.RemoveS {
		if err := unsetSpecialBits(&ctr, findCmdFactory, s.Security.SpecialFiles.Excludes); err != nil {
			return fmt.Errorf("removing SUID and SGID bits from files: %w", err)
		}
		logger.Debugln("command to remove SUID and SGID bits from files ran successfully")
	}

	if options.Digest != "" {
		s.Config.Annotations["org.github.ok-ryoko.turret.spec.digest"] = options.Digest
	}

	ports := make([]string, len(s.Config.Ports))
	for i, p := range s.Config.Ports {
		ports[i] = p.String()
	}

	configureOptions := configureOptions{
		clearAnnotations: s.Config.Clear.Annotations,
		annotations:      s.Config.Annotations,
		clearAuthor:      s.Config.Clear.Author,
		author:           s.Config.Author,
		clearCommand:     s.Config.Clear.Command,
		command:          s.Config.Command,
		createdBy:        s.Config.CreatedBy,
		clearEntrypoint:  s.Config.Clear.Entrypoint,
		entrypoint:       s.Config.Entrypoint,
		clearEnvironment: s.Config.Clear.Environment,
		environment:      s.Config.Environment,
		clearLabels:      s.Config.Clear.Labels,
		labels:           s.Config.Labels,
		clearPorts:       s.Config.Clear.Ports,
		ports:            ports,
		workDir:          s.Config.WorkDir,
	}
	if s.User != nil {
		configureOptions.user = s.User.Name
	}
	configure(&ctr, configureOptions)
	logger.Debugln("configured image")

	logger.Debugln("committing image...")
	commitOptions := commitOptions{
		keepHistory: s.This.KeepHistory,
		latest:      options.Latest,
	}
	imageID, err := commit(
		&ctr,
		ctx,
		store,
		s.This.Repository,
		s.This.Tag,
		commitOptions,
	)
	if err != nil {
		return fmt.Errorf("committing image: %w", err)
	}
	logger.Infoln(imageID)

	return nil
}

// ExecuteOptions holds options for the build pipeline.
type ExecuteOptions struct {
	// SHA256 digest of the spec file to apply as an annotation to the new image
	Digest string

	// Overwrite the target image if it already exists
	Force bool

	// Keep the working container (even in the event of an error)
	Keep bool

	// Ensure the creation of the `latest` tag
	Latest bool

	// Log the standard output of container processes
	LogCommands bool

	// Retrieve the image only if it's not already in local storage
	Pull bool
}

// cleanPackageCaches cleans the package caches in the working container.
func cleanPackageCaches(c *container.Container, p container.PackageFrontendInterface) error {
	if err := p.CleanCaches(c); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// commit commits an image from the working container to storage and returns
// the ID of the newly created image, assuming `repository` and `tag` are
// nonempty strings from which a valid image reference can be composed.
func commit(
	c *container.Container,
	ctx context.Context,
	store storage.Store,
	repository string,
	tag string,
	options commitOptions,
) (string, error) {
	co := buildah.CommitOptions{
		PreferredManifestType: manifestType,
		Compression:           archive.Gzip,
		HistoryTimestamp:      &time.Time{},
		OmitHistory:           false,
		Squash:                true,
	}

	if options.latest && tag != "latest" {
		co.AdditionalTags = append(
			co.AdditionalTags,
			fmt.Sprintf("%s:latest", repository),
		)
	}

	if options.keepHistory {
		co.HistoryTimestamp = nil
		co.OmitHistory = false
	}

	imageRef := fmt.Sprintf("%s:%s", repository, tag)
	storageRef, err := is.Transport.ParseStoreReference(store, imageRef)
	if err != nil {
		return "", fmt.Errorf("parsing reference: %w", err)
	}
	imageID, _, _, err := c.Builder.Commit(ctx, storageRef, co)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	return imageID, nil
}

// commitOptions holds options for committing an image from the working
// container to storage.
type commitOptions struct {
	// Preserve the image history and timestamps of the files in the working
	// container's file system
	keepHistory bool

	// Ensure that the `latest` tag is created
	latest bool
}

// configure alters the metadata on and execution of the working container.
func configure(c *container.Container, options configureOptions) {
	if options.clearAnnotations {
		for k := range c.Builder.Annotations() {
			if !strings.HasPrefix("org.opencontainers.image.base", k) {
				c.Builder.UnsetAnnotation(k)
			}
		}
	}
	for k, v := range options.annotations {
		c.Builder.SetAnnotation(k, v)
	}

	if options.clearAuthor {
		c.Builder.SetMaintainer("")
	}
	if options.author != "" {
		c.Builder.SetMaintainer(options.author)
	}

	if options.clearCommand {
		c.Builder.SetCmd([]string{})
	}
	if len(options.command) > 0 {
		c.Builder.SetCmd(options.command)
	}

	if options.createdBy != "" {
		c.Builder.SetCreatedBy(options.createdBy)
	}

	if options.clearEntrypoint {
		c.Builder.SetEntrypoint([]string{})
	}
	if len(options.entrypoint) > 0 {
		c.Builder.SetEntrypoint(options.entrypoint)
	}

	if options.clearEnvironment {
		c.Builder.ClearEnv()
	}
	for k, v := range options.environment {
		c.Builder.SetEnv(k, v)
	}

	if options.clearLabels {
		c.Builder.ClearLabels()
	}
	for k, v := range options.labels {
		c.Builder.SetLabel(k, v)
	}

	c.Builder.SetOS("linux")

	if options.clearPorts {
		c.Builder.ClearPorts()
	}
	for _, p := range options.ports {
		c.Builder.SetPort(p)
	}

	if options.workDir != "" {
		c.Builder.SetWorkDir(options.workDir)
	}

	if options.user != "" {
		c.Builder.SetUser(options.user)
	}
}

// configureOptions holds configuration options for the working container.
type configureOptions struct {
	// Clear all annotations inherited from the base image
	clearAnnotations bool

	// Set or update one or more annotations
	annotations map[string]string

	// Clear the author inherited from the base image
	clearAuthor bool

	// Provide contact information for the image maintainer
	author string

	// Clear the command inherited from the base image
	clearCommand bool

	// Set the default command (or the parameters, if an entrypoint is set)
	command []string

	// Describe how the image was built
	createdBy string

	// Clear the entrypoint inherited from the base image
	clearEntrypoint bool

	// Set the entrypoint
	entrypoint []string

	// Unset all environment variables inherited from the base image
	clearEnvironment bool

	// Set or update one or more environment variables
	environment map[string]string

	// Clear all labels inherited from the base image
	clearLabels bool

	// Set or update one or more labels
	labels map[string]string

	// Close all exposed network ports inherited from the base image
	clearPorts bool

	// Expose one or more network ports
	ports []string

	// Set the user as whom the entrypoint or command should run
	user string

	// Set the default directory in which the entrypoint or command should run
	workDir string
}

// copyFiles copies one or more files on the host's file system to the working
// container's file system, assuming `base` and `dest` are absolute file paths
// and `srcs` is a nonempty slice of file paths.
//
// `base` is an absolute path to a directory on the host's file system against
// which relative paths in `srcs` should be resolved.
//
// `dest` is an absolute path to a destination on the working container's file
// system. If the destination ends with a path separator, then it's assumed to
// be a directory.
//
// `srcs` is a slice of relative or absolute paths to items on the host's file
// system. Relative paths are resolved with respect to `base`.
//
// If there is only one source item and the destination does not end with a
// path separator, then copy the item to the parent directory in the
// destination, renaming the item to match the destination as needed.
func copyFiles(c *container.Container, base string, dest string, srcs []string, options copyFilesOptions) error {
	patterns := make([]string, len(srcs))
	for i, s := range srcs {
		patterns[i] = fmt.Sprintf("!%s", s)
	}
	excludes := append([]string{"*"}, patterns...)
	if len(options.excludes) > 0 {
		excludes = append(excludes, options.excludes...)
	}

	aco := buildah.AddAndCopyOptions{
		ContextDir: base,
		Excludes:   excludes,
	}

	if options.owner != "" {
		aco.Chown = options.owner
	}

	if options.mode != 0 {
		aco.Chmod = fmt.Sprint(options.mode)
	}

	if options.removeSpecial {
		aco.StripSetuidBit = true
		aco.StripSetgidBit = true
	}

	if err := c.Builder.Add(dest, false, aco, base); err != nil {
		return fmt.Errorf("copying files from %q to %q: %w", base, dest, err)
	}

	return nil
}

// copyFilesOptions holds options for copying files from the host's file system
// to the working container's file system.
type copyFilesOptions struct {
	// Source files in the base directory to exclude from the copy operation;
	// may contain gitignore-style glob patterns
	excludes []string

	// Set the mode of the copied files to this integer
	mode uint32

	// Transfer ownership of the copied files to this user
	owner string

	// Remove all SUID and SGID bits from the files copied to the working
	// container
	removeSpecial bool
}

// createUser creates the sole unprivileged user of the working container,
// assuming `name` is a nonempty string.
func createUser(c *container.Container, u container.UserFrontendInterface, name string, options user.Options) error {
	if err := u.CreateUser(c, name, options); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// installPackages installs one or more packages to the working container.
func installPackages(c *container.Container, p container.PackageFrontendInterface, packages []string) error {
	if err := p.Install(c, packages); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// unsetSpecialBits removes the SUID and SGID bits from files in the working
// container, assuming the availability of the chmod and find core utilities
// and searching only real (non-device) file systems.
//
// `excludes` is a slice of absolute paths to real files in the working
// container for which to keep SUID and SGID bits.
func unsetSpecialBits(c *container.Container, f find.CommandFactory, excludes []string) error {
	var targets []string

	{
		cmd, capabilities := f.NewFindSpecialCmd()
		ro := c.DefaultRunOptions()
		ro.AddCapabilities = capabilities
		outText, errText, err := c.Run(cmd, ro)
		if err != nil {
			errContext := "searching for special files"
			if errText != "" {
				errContext = fmt.Sprintf("%s (%q)", errContext, errText)
			}
			return fmt.Errorf("%s: %w", errContext, err)
		}
		if len(outText) > 0 {
			targets = strings.Split(strings.ReplaceAll(strings.TrimSpace(outText), "\r\n", "\n"), "\n")
		}
	}

	if len(excludes) > 0 {
		excludeSet := map[string]bool{}
		for _, e := range excludes {
			excludeSet[e] = true
		}

		var filteredTargets []string
		for _, t := range targets {
			if _, ok := excludeSet[t]; !ok {
				filteredTargets = append(filteredTargets, t)
			}
		}

		targets = filteredTargets
	}

	if len(targets) > 0 {
		cmd := append([]string{"chmod", "-s"}, targets...)

		// CAP_FSETID is a member of the chmod effective capability set but is
		// neither sufficient nor necessary for this operation
		//
		ro := c.DefaultRunOptions()
		ro.AddCapabilities = []string{
			"CAP_DAC_READ_SEARCH",
			"CAP_FOWNER",
		}

		_, errText, err := c.Run(cmd, ro)
		if err != nil {
			errContext := "unsetting special bit"
			if errText != "" {
				errContext = fmt.Sprintf("%s (%q)", errContext, errText)
			}
			return fmt.Errorf("%s: %w", errContext, err)
		}
	}

	return nil
}

// upgradePackages upgrades the packages in the working container.
func upgradePackages(c *container.Container, p container.PackageFrontendInterface) error {
	if err := p.Upgrade(c); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}
