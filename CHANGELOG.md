# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Add `global.podSecurityStandards.enforced` value for PSS migration.

### Changed

- Update `controller-gen` to 0.10.0.

## [0.6.0] - 2022-10-04

### Changed

- `PodSecurityPolicy` are removed on newer k8s versions, so only apply it if object is registered in the k8s API.

## [0.5.2] - 2022-07-01

### Removed

- Remove ingress registrar, as ingress DNS record will be created by external-dns.

## [0.5.1] - 2022-06-22

### Fixed

- Skip deletion when the zone is deleted.

## [0.5.0] - 2022-06-22

### Added

- Add DNS record for bastion nodes.

## [0.4.0] - 2022-06-02

### Changed

- Ignore non LoadBalancer services when registering ingress record. The nginx ingress app installs multiple ClusterIP services.

## [0.3.0] - 2022-06-01

### Changed

- Improve logging by adding gcp cluster name being reconciled.
- Add log to mark start and end of reconcile
- Use microerror consistently everywhere

## [0.2.0] - 2022-05-05

## [0.1.0] - 2022-05-05

[Unreleased]: https://github.com/giantswarm/dns-operator-gcp/compare/v0.6.0...HEAD
[0.6.0]: https://github.com/giantswarm/dns-operator-gcp/compare/v0.5.2...v0.6.0
[0.5.2]: https://github.com/giantswarm/dns-operator-gcp/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/giantswarm/dns-operator-gcp/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/giantswarm/dns-operator-gcp/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/giantswarm/dns-operator-gcp/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/giantswarm/dns-operator-gcp/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/giantswarm/dns-operator-gcp/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/giantswarm/dns-operator-gcp/releases/tag/v0.1.0
