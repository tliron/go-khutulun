*This is an early release. Some features are not yet fully implemented.*

Khutulun
========

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Latest Release](https://img.shields.io/github/release/tliron/khutulun.svg)](https://github.com/tliron/khutulun/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/tliron/khutulun)](https://goreportcard.com/report/github.com/tliron/khutulun)

A service orchestrator for machine clusters that speaks
[TOSCA](https://www.oasis-open.org/committees/tosca/).

Khutulun addresses similar uses as [Kubernetes](https://kubernetes.io/) does and can provide an
alternative, more straightforward solution.

Its primary design goal is that the outcome of orchestration would be no different from what a
sysadmin would do themselves. If you want to simply install and run a bare process on a machine,
Khutulun will do that for you. If you want straightforward networking based on reserved TCP ports,
Khutulun won't do anything more than keep track of those ports for you. More complex deployments
using containers, virtual machines, and virtual networks are also supported. Khutulun's aim is to
manage complexity without getting in the way of simplicity.

The guiding paradigm is policy-driven service composition based on graph representations of the
service's topology. Khutulun provides custom discovery that allows components to find each other,
form a service mesh, and modify the topologies to which they belong ("Day 2" and the "operator
pattern"). There are tools for injecting discovered data into configuration files and environment
variables to allow "legacy" applications to participate in the mesh.

Plugins
-------

Khutulun is modular and extensible. Resource types are handled by a cooperative ecosystem of plugins,
the main plugin types being for running compute workloads, for networking, and for storage. Plugins
can call other plugins and can themselves be implemented as workloads on the cluster.

Some included resource types and plugins:

* Bare processes: self-contained or otherwise installable executables and scripts
* Containers or pods of containers using [Podman](https://podman.io/)
* Pristine containers using [Distrobox](https://distrobox.privatedns.org/) (on top of Podman)
* System containers using [systemd-nspawn](https://www.freedesktop.org/software/systemd/man/systemd-nspawn.html)
* Virtual machines using [libvirt](https://libvirt.org/)
* TCP port reservation with support for exposure through [Firewalld](https://firewalld.org/)
* Local or networked directory storage

Plugins can optionally wrap resources in usermode systemd units. This provides a unified admin
experience as well as resilience in the case of failures and restarts.

Clusters
--------

Cluster formation is emergent and based on the [SWIM gossip protocol](https://ieeexplore.ieee.org/document/1028914).
At the minimum you need just one "seed" host to bootstrap it, but because all hosts are "masters"
the cluster can survive with as little as one arbitrary host.

Khutulun doesn't distribute its management state among hosts. Instead it simply requires that all
hosts have access to the same shared filesystem. A simple NFS share should be enough even for large
clusters. Coordination is handled via [flock](https://man7.org/linux/man-pages/man2/flock.2.html).

What about setting up the cluster hosts? Bare metal tasks like partitioning drives, installing
operating systems, and configuring networking and other essential services? Or cloud tasks like
provisioning virtual machines, virtual storage, and virtual networks? Simply put, it's out of the
scope of Khutulun. Use a dedicated infrastructure manager instead. Khutulun can interact with such
tools, for example to allow workloads to modify their own cluster, or to use a Khutulun cluster as
a dedicated "management cluster" that, well, manages the hardware of all other clusters. Included is
a plugin for [Terraform](https://www.terraform.io/) that allows for that.

Get It
------

[![Download](assets/media/download.png "Download")](https://github.com/tliron/khutulun/releases)

FAQ
---

### Why TOSCA?

TOSCA is an open standard with broad industry support. It is, as of version 2.0, a pure
object-oriented language that relies on "profiles", or type libraries, that can work with plugins
to provide specific implementations. Khutulun comes with its own TOSCA profile and ecosystem of
plugins. You are encouraged to add your own.

One of the hallmarks of TOSCA is that every service is a topological graph. Moreover, the edges
of the graph are 1st-class citizens. This killer feature supercharges your modeling power for the
cloud.

The developers of Khutulun are involved in the TOSCA community and committed to improving the
standard.

### Why support bare processes? Don't containers provide better isolation?

Yes, containers indeed provide better isolation and Khutulun supports them out the box via Podman,
Distrobox (on top of Podman), and systemd-nspawn.

But don't just jump on the bandwagon, ask yourself: Is isolation really what you need for your use
case? And do you understand and are willing to pay for what it costs? We are in the midst of an
architectural shift towards service composition and away from component isolation. Isolation is often
beneficial, and in some specific use cases even necessary, but if isolation technologies get in the way
of collaboration technologies then you're are shooting yourself in the foot. Most container
technologies require you to build ready-to-run container images and stand up container image registries
to store them, adding significant complexity to your development and deployment workflows. Also complex
is managing container networking across clusters. If you entirely own your cluster and workloads then
it might save you mountains of pain to simply use bare processes with bare networking.

Consider Distrobox as a Goldilocks solution: it provides pristine containers that provide only a
minimal operating system but no workloads, so you can run your workloads there instead of on the
bare host. Khutulun will handle the heavy lifting for you. The result may give you the best of both
worlds.

### Why not use a distributed key-value store like [etcd](https://etcd.io/) for management state?

What's wrong with just having a filesystem shared among all hosts? Seriously, why make things more
complicated than they have to be?

Also note that etcd has strict limits on the size of documents, which is an obstacle for sharing large,
useful binary artifacts. That means that if you need to share large, useful binary artifacts you will
need to deploy yet another system. Are we winning yet?

### Why is there no custom Khutulun cluster installer? Why recommend using Terraform and other tools instead?

Infrastructure management is a solved problem. Let's please not reinvent the wheel just for Khutulun
to have its own opinion.

### Why is it called "Khutulun"?

[Khutulun](https://en.wikipedia.org/wiki/Khutulun) (Mongolian: Хотулун) was a fabled Mongolian warrior,
daughter of Kublai Khan's cousin, Kaidu.

She was likely the inspiration for *Turandot* (Persian: Turandokht), the protagonist of
[Giacomo Puccini](https://en.wikipedia.org/wiki/Giacomo_Puccini)'s
[opera of the same name](https://en.wikipedia.org/wiki/Turandot), which was in turn likely inspired
by Count Carlo Gozzi's *commedia dell'arte*
[play of the same name](https://en.wikipedia.org/wiki/turandot_(Gozzi)).

And [Puccini](https://puccini.cloud/) is the TOSCA processor that drives Khutulun.

### How do I pronounce "Khutulun"?

* International level: "KOO-too-loon"
* Cosmopolitan level: "XOO-too-loon" (the "x" sounds like "ch" in "Johann Sebastian Bach")
* Expert level: Mongolian "Хотулун" ([video](https://www.youtube.com/watch?v=uP0BagZ-ZCE&t=58s))
