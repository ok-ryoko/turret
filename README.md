<p align="center">
  <img
    src="./docs/img/icon.svg"
    title="Turret"
    alt="minimalistic and isometric view of a red–brown castle turret superposed onto a clear blue sky during the day"
    height="128"
  >
</p>

<h2 align="center">Turret</h2>

Turret is a command-line utility that produces rootless [OCI] images of independent [GNU/Linux] distributions from [TOML] specification files (specs). Turret is also a high-level interface over [Buildah].

## Compatible GNU/Linux distros

- [Alpine]
- [Arch]
- [Debian]
- [Fedora]
- [openSUSE]
- [Void]

## Getting Turret

Turret is pre-release software and must be built from source.

### Requirements

- [Go][download and install Go] 1.20 or newer
- [GNU Compiler Collection] (GCC)
- [pkg-config]
- [Btrfs] library development headers (libbtrfs)
- [Linux Kernel Device Mapper] library development headers (libdevmapper)
- [GnuPG Made Easy] library development headers (libgpgme)

#### Alpine

> *Tested on [docker.io/library/alpine:3.18.0]*

```sh
apk add \
    btrfs-progs-dev \
    gcc \
    go \
    gpgme-dev \
    lvm2-dev \
    pkgconf
```

#### Arch

> *Tested on [docker.io/library/archlinux:base-20230514.0.150299]*

```sh
pacman -Sy \
    btrfs-progs \
    device-mapper \
    gcc \
    go \
    gpgme \
    pkgconf
```

#### Debian

> *Tested on [docker.io/library/golang:1.20.4-bullseye]*

```sh
apt update
apt install -y \
    libbtrfs-dev \
    libdevmapper-dev \
    libgpgme11-dev
```

#### Fedora

> *Tested on [registry.fedoraproject.org/fedora:38-x86_64]*

```sh
dnf -y install \
    btrfs-progs-devel \
    device-mapper-devel \
    gcc \
    golang \
    gpgme-devel \
    pkgconf
```

#### openSUSE

> *Tested on [registry.opensuse.org/opensuse/leap:15.5]*

```sh
zypper in -y \
    device-mapper-devel \
    gcc \
    go1.20 \
    libbtrfs-devel \
    libgpgme-devel \
    pkg-config
```

#### Void

> *Tested on [ghcr.io/void-linux/void-linux:20230204RC01-full-x86_64]*

```sh
xbps-install -Syu
xbps-install -Sy \
    device-mapper-devel \
    gcc \
    go \
    gpgme-devel \
    libbtrfs-devel \
    pkg-config
```

### Instructions

To try Turret out:

```sh
go install github.com/ok-ryoko/turret/cmd/turret@latest
```

To work on Turret:

```sh
git clone https://github.com/ok-ryoko/turret
cd turret
make build
ls -hl ./cmd/turret/build/turret
```

## Building OCI images with Turret

Create a spec, e.g., *example.toml*, containing the following data:

```toml
distro = "fedora"
repository = "turret/f38-dev"
tag = "0.1.0"

upgrade = true
packages = ["fish", "gcc", "git", "neovim"]
clean = true

[from]
repository = "registry.fedoraproject.org/fedora"
tag = "38-x86_64"

[user]
create = true
name = "fuser"
login-shell = "fish"
```

This spec defines a minimal development environment. Let’s feed it to Turret:

```sh
turret build -flv 3 ./example.toml
```

```
time="2023-05-17T19:43:49Z" level=debug msg="processed spec path"
time="2023-05-17T19:43:49Z" level=debug msg="created build spec"
time="2023-05-17T19:43:50Z" level=debug msg="created working container from image 'registry.fedoraproject.org/fedora:38-x86_64'"
time="2023-05-17T19:43:50Z" level=debug msg="created Fedora Linux Turret builder"
time="2023-05-17T19:43:50Z" level=debug msg="upgrading packages..."
time="2023-05-17T19:46:30Z" level=debug msg="package installation step succeeded"
time="2023-05-17T19:46:31Z" level=debug msg="package cache cleaning step succeeded"
time="2023-05-17T19:46:31Z" level=debug msg="created unprivileged user 'fuser'"
time="2023-05-17T19:46:31Z" level=debug msg="file copy step succeeded"
time="2023-05-17T19:46:31Z" level=debug msg="configured image"
time="2023-05-17T19:46:31Z" level=debug msg="committing image..."
time="2023-05-17T19:46:51Z" level=info msg="built and committed Fedora Linux image e0a7f983b32fd95824b6e574a68466c7d5c76f0c304690c37686dbd5eb1c8b5a"
```

… where we:

- force-overwrote the image if it already existed in the local registry (`-f`);
- created or updated the `latest` tag for the image (`-l`), and
- printed debug output from Turret but not any container processes (`-v 3`).

> Turret won’t dereference symbolic links and expects the spec to be resolvable with respect to both your home and current working directories.

Here are some properties of the image we just built:

- Can be referenced as *localhost/turret/f38-dev:0.1.0* or *localhost/turret/f38-dev:latest*
- Runs with no capabilities
- Contains the [fish shell], the [GNU Compiler Collection], [Git] and [Neovim] plus their dependencies
- Contains `fuser`, an unprivileged user
- Executes the process `/bin/sh -c /usr/bin/fish` as `fuser` in */home/fuser*

## Running Turret containers

### Requirements

Install the following packages on the host:

- podman
- fuse-overlayfs
- shadow-utils or newuid
- slirp4netns

The last three are needed to run Podman in rootless mode.

### Example

You can run Turret containers using Podman. Continuing the previous example:

```console
$ podman run -it --rm localhost/turret/f38-dev:0.1.0
Welcome to fish, the friendly interactive shell
Type help for instructions on how to use fish
fuser@60d55ac61497 ~> cat /usr/lib/fedora-release
Fedora release 38 (Thirty Eight)
```

### Configuring containers

Turret allows us to define images as structured data. Can we do so for container execution? Yes: We can write an image-specific [*containers.conf*][containers.conf] file and set the `CONTAINERS_CONF_OVERRIDE` environment variable accordingly:

```sh
CONTAINERS_CONF_OVERRIDE=./image.conf podman run -it image
```

## Undefined behavior

- Running Turret as a privileged user
- Specifying a distro that doesn’t match the distro in the base image

## Community

### Understanding our code of conduct

Please take time to read [our code of conduct] before reaching out for support or making a contribution.

### Getting support

If you’re encountering unexpected or undesirable program behavior, check the [issue tracker] to see whether your problem has already been reported. If not, please consider taking time to create a bug report.

If you have questions about using the program or participating in the community around the program, consider [starting a discussion][discussions].

Please allow up to 1 week for a maintainer to reply to an issue or a discussion.

## When should I not use Turret?

Turret may not be right for you if you require:

- a stable, well-tested or well-documented build engine for use in production;
- access to hardware (e.g., to render graphical applications, access storage, etc.);
- complex image configuration;
- nested virtualization, or
- a greater degree of isolation than containers can provide.

Consider using [Distrobox], [Toolbox] or a dedicated virtual or physical machine instead.

## See also

- [Buildah]
- [Podman]
- [Distrobox]
- [Toolbox]

## License

Turret is free and open source software licensed under the [Apache 2.0 license].

## Acknowledgements

The Turret logo was made in [Inkscape].

Turret builds over software published by the [Containers] organization, some members of which are sponsored by [Red Hat].

The following resources have been instrumental in preparing this repository for community contributions:

- [Open Source Guides]
- [the GitHub documentation][GitHub documentation] and [the github/docs repository][github/docs]
- [the tokio contributing guidelines][tokio contributing guidelines]

[Alpine]: https://www.alpinelinux.org
[Apache 2.0 license]: ./LICENSE
[Arch]: https://archlinux.org
[Btrfs]: https://wiki.archlinux.org/title/Btrfs
[Buildah]: https://github.com/containers/buildah
[containers.conf]: https://github.com/containers/common/blob/main/docs/containers.conf.5.md
[Containers]: https://github.com/containers
[Debian]: https://www.debian.org
[discussions]: https://github.com/ok-ryoko/turret/discussions
[Distrobox]: https://github.com/89luca89/distrobox
[download and install Go]: https://go.dev/doc/install
[Fedora]: https://www.fedoraproject.org
[fish shell]: https://fishshell.com
[Git]: https://git-scm.com
[GitHub documentation]: https://docs.github.com/en
[github/docs]: https://github.com/github/docs
[GNU Compiler Collection]: https://gcc.gnu.org
[GNU/Linux]: https://www.gnu.org/
[GnuPG Made Easy]: https://git.gnupg.org/cgi-bin/gitweb.cgi?p=gpgme.git
[Inkscape]: https://inkscape.org/
[issue tracker]: https://github.com/ok-ryoko/turret/issues
[Linux Kernel Device Mapper]: https://sourceware.org/dm/
[Neovim]: https://neovim.io
[OCI]: https://opencontainers.org/
[Open Source Guides]: https://opensource.guide/
[openSUSE]: https://www.opensuse.org
[our code of conduct]: ./CODE_OF_CONDUCT.md
[pkg-config]: https://www.freedesktop.org/wiki/Software/pkg-config/
[Podman]: https://github.com/containers/podman
[Red Hat]: https://redhatofficial.github.io/#!/main
[tokio contributing guidelines]: https://github.com/tokio-rs/tokio/blob/d7d5d05333f7970c2d75bfb20371450b5ad838d7/CONTRIBUTING.md
[TOML]: https://toml.io/
[Toolbox]: https://github.com/containers/toolbox
[Void]: https://voidlinux.org

[docker.io/library/alpine:3.18.0]: https://hub.docker.com/layers/library/alpine/3.18.0/images/sha256-c0669ef34cdc14332c0f1ab0c2c01acb91d96014b172f1a76f3a39e63d1f0bda?context=explore
[docker.io/library/archlinux:base-20230514.0.150299]: https://hub.docker.com/layers/library/archlinux/base-20230514.0.150299/images/sha256-f081f7f60b83cfeaff651e4ca03e4d23bf6ce6a5045594ea9b983aa686acb817?context=explore
[docker.io/library/golang:1.20.4-bullseye]: https://hub.docker.com/layers/library/golang/1.20.4-bullseye/images/sha256-5099ad46335916ab90a4ce5ead4e01cb6eefc2f0296ef9f04af61b3e60f96c78?context=explore
[ghcr.io/void-linux/void-linux:20230204RC01-full-x86_64]: https://github.com/void-linux/void-docker/pkgs/container/void-linux/68157358?tag=20230204RC01-full-x86_64
[registry.fedoraproject.org/fedora:38-x86_64]: https://registry.fedoraproject.org/repo/fedora/tags/
[registry.opensuse.org/opensuse/leap:15.5]: https://build.opensuse.org/package/show/openSUSE:Containers:Leap/15.5
