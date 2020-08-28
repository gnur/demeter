# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

### [2.4.2](https://github.com/gnur/demeter/compare/v2.4.1...v2.4.2) (2020-08-28)

### [2.4.1](https://github.com/gnur/demeter/compare/v2.4.0...v2.4.1) (2020-08-28)

## [2.4.0](https://github.com/gnur/demeter/compare/v2.3.0...v2.4.0) (2020-08-28)


### Features

* **file format:** Allow filtering other file formats then epub ([16fce3a](https://github.com/gnur/demeter/commit/16fce3aafe3ab4db465c88cdd90135e2e798f513)), closes [#3](https://github.com/gnur/demeter/issues/3)

## [2.3.0](https://github.com/gnur/demeter/compare/v2.2.0...v2.3.0) (2019-11-11)


### Features

* **Add host enable-all command:** demeter host enable-all re-enables all hosts ([be807ae](https://github.com/gnur/demeter/commit/be807aedfe7019dde9018c5b899a7e59a107f968))
* Hosts don't get disabled if the last run succeeded ([2651241](https://github.com/gnur/demeter/commit/26512413435174aad101ddf9712bca1659fd60fa))


### Bug Fixes

* Added the enable all command to options ([0138c33](https://github.com/gnur/demeter/commit/0138c332584631e79fc36f67fb4fbd9b280a1810))

## [2.2.0](https://github.com/gnur/demeter/compare/v2.1.0...v2.2.0) (2019-11-11)


### Features

* **info:** Added version subcommand ([0c3d8fd](https://github.com/gnur/demeter/commit/0c3d8fd891d9e47b09168aaf59cb3396911fbe14))

## [2.1.0](https://github.com/gnur/demeter/compare/v2.0.0...v2.1.0) (2019-11-08)


### Features

* **hosts:** Allow multiple hosts to be added at once ([5e00138](https://github.com/gnur/demeter/commit/5e00138858e1fb60ec867731faac9bda2c08bf19)), closes [#1](https://github.com/gnur/demeter/issues/1)


### Bug Fixes

* Improve error message on timeout ([41fbb50](https://github.com/gnur/demeter/commit/41fbb501d8df95680a8bd28f6adf44d18589956b)), closes [#2](https://github.com/gnur/demeter/issues/2)

## [2.0.0](https://github.com/gnur/demeter/compare/v1.0.3...v2.0.0) (2019-11-05)


### ⚠ BREAKING CHANGES

* Removed runv2 command

### Features

* Removed old run and replaced it with runv2 ([8b5fa0e](https://github.com/gnur/demeter/commit/8b5fa0e39d294cd04a4216cc70c5d748c13891f0))

### [1.0.3](https://github.com/gnur/demeter/compare/v1.0.2...v1.0.3) (2019-11-02)

### [1.0.2](https://github.com/gnur/demeter/compare/v1.0.1...v1.0.2) (2019-11-02)

### [1.0.1](https://github.com/gnur/demeter/compare/v1.0.0...v1.0.1) (2019-11-02)

## 1.0.0 (2019-11-02)


### Features

* Improved logging ([ab7c4f2](https://github.com/gnur/demeter/commit/ab7c4f23a336e04fbfdf589dedc747588f4664f2))
* Made stuff run more in parallel ([d0d8700](https://github.com/gnur/demeter/commit/d0d8700306a7761ed1b3b0b73e48949ded43c1b1))
* **runv2:** added new runv2 command for better performance ([72a62fc](https://github.com/gnur/demeter/commit/72a62fcf615601675764eed0818a5cca070f7c03))


### Bug Fixes

* Limiting runs to twice a day per host ([f2c4c14](https://github.com/gnur/demeter/commit/f2c4c14d2a511fa9d6cc67c9427e6fc6d11a35b4))
* not disabling hosts and reduced redundant calls ([27025d6](https://github.com/gnur/demeter/commit/27025d6c84ca865a6f989394463804c0a6c56a8f))
* Some more info per host and fix out of range error in bookindb ([4e8cd84](https://github.com/gnur/demeter/commit/4e8cd84fd464b688c0a418921f1a20cc6d676441))
