<p align="center">
<img src="image/module-sdk-small-logo.png" alt="module-sdk logo" />
</p>

<p align="center">
<a href="https://hub.docker.com/r/flant/module-sdk"><img src="https://img.shields.io/docker/pulls/flant/module-sdk.svg?logo=docker" alt="docker pull flant/module-sdk"/></a>
 <a href="https://github.com/flant/module-sdk/discussions"><img src="https://img.shields.io/badge/GitHub-discussions-brightgreen" alt="GH Discussions"/></a>
</p>

**Module-sdk** is a toolkit for developing modules for Kubernetes clusters.

This SDK is not a module for a _particular software product_ but rather a framework to build modules that can be integrated into Kubernetes clusters by Deckhouse. Module-sdk provides developers with tools and patterns to create standardized, maintainable, and extensible Kubernetes modules.

Module-sdk serves as the foundation for creating consistent module implementations that follow best practices for Kubernetes integration.

Module-sdk provides:

- __Simplified module development__: use standardized patterns and tools familiar to developers. Compatible with Go.
- __Kubernetes integration primitives__: easily define resources, hooks, and configuration required by your module.
- __Testing frameworks__: comprehensive testing utilities for ensuring module reliability.

# Community

Please feel free to reach developers/maintainers and users via [GitHub Issues][issues] for any questions regarding module-sdk.

You're also welcome to follow [@flant_com][twitter] to stay informed about all our Open Source initiatives.

# License

Apache License 2.0, see [LICENSE][license].

[issues]: https://github.com/deckhouse/module-sdk/issues
[license]: https://github.com/deckhouse/module-sdk/blob/main/LICENSE
[twitter]: https://twitter.com/flant_com
