# Changelog

## [2.1.0](https://github.com/silence-operator/silence-operator/compare/v2.0.1...v2.1.0) (2025-08-04)


### Features

* **crd:** Add defaults for bool values ([180f284](https://github.com/silence-operator/silence-operator/commit/180f284a3d7b6c0959268bd8cf651df839479a59))
* **crd:** Add defaults for bool values ([8f6a5f7](https://github.com/silence-operator/silence-operator/commit/8f6a5f78519b357a45adbe428729da857933d70e))

## [2.0.1](https://github.com/silence-operator/silence-operator/compare/v2.0.0...v2.0.1) (2025-07-24)


### Bug Fixes

* **helm:** Add protocol to default url for alertmanager ([add4389](https://github.com/silence-operator/silence-operator/commit/add43891a568b20ad96f85f48281123602b4bae7))

## [2.0.0](https://github.com/silence-operator/silence-operator/compare/v1.1.6...v2.0.0) (2025-07-23)


### ⚠ BREAKING CHANGES

* Provide alertmanager URL instead of host

### Features

* Provide alertmanager URL instead of host ([20faf8f](https://github.com/silence-operator/silence-operator/commit/20faf8f98be763395e77864d1e940c6212a6aeda))


### Bug Fixes

* **controller:** Do not use deprecated requeue option ([6ebda0d](https://github.com/silence-operator/silence-operator/commit/6ebda0dcd6cfefbed90d15567de270b41cb8d827))

## [1.1.6](https://github.com/silence-operator/silence-operator/compare/v1.1.5...v1.1.6) (2025-07-23)


### Bug Fixes

* **operator:** Fix matcher to string conversion ([9b7fdc8](https://github.com/silence-operator/silence-operator/commit/9b7fdc833dbc956a920f617a953f60bdb9a6b6f5))

## [1.1.5](https://github.com/silence-operator/silence-operator/compare/v1.1.4...v1.1.5) (2025-07-23)


### Bug Fixes

* **helm:** Remove manager command from deployment ([fa6c871](https://github.com/silence-operator/silence-operator/commit/fa6c8713424b94588fb49d351b4f6be9186c0225))

## [1.1.4](https://github.com/silence-operator/silence-operator/compare/v1.1.3...v1.1.4) (2025-07-23)


### Bug Fixes

* **ci:** Use goreleaser version for helm ([82894a2](https://github.com/silence-operator/silence-operator/commit/82894a24b81745e6a31d58d98231310633c22ddd))

## [1.1.3](https://github.com/silence-operator/silence-operator/compare/v1.1.2...v1.1.3) (2025-07-23)


### Bug Fixes

* **helm:** Remove "v" prefix for image tag ([47b3ef4](https://github.com/silence-operator/silence-operator/commit/47b3ef470844d349b8d2ebdbef8918e8403a02f9))

## [1.1.2](https://github.com/silence-operator/silence-operator/compare/v1.1.1...v1.1.2) (2025-07-23)


### Bug Fixes

* **ci:** Merge image and helm workflows ([9a6d21f](https://github.com/silence-operator/silence-operator/commit/9a6d21fe52fae228deee06696daef24db0fb06b2))

## [1.1.1](https://github.com/silence-operator/silence-operator/compare/v1.1.0...v1.1.1) (2025-07-23)


### Bug Fixes

* **ci:** Update chart version during chart release ([494aab3](https://github.com/silence-operator/silence-operator/commit/494aab39cac2e78f018a40f45d3be17eec136fc5))

## [1.1.0](https://github.com/silence-operator/silence-operator/compare/v1.0.4...v1.1.0) (2025-07-23)


### Features

* **ci:** Add releases for helm ([012f33f](https://github.com/silence-operator/silence-operator/commit/012f33f6df2894976023ae791d554092184acd67))
* **ci:** Add releases for helm ([b127eed](https://github.com/silence-operator/silence-operator/commit/b127eed903d5cb28fbeaacd6ea4d043687bb87d1))

## [1.0.4](https://github.com/silence-operator/silence-operator/compare/v1.0.3...v1.0.4) (2025-07-23)


### Bug Fixes

* **ci:** Change token for goreleaser ([633c817](https://github.com/silence-operator/silence-operator/commit/633c81705beaa8e16afe5b1aa2a8d6438198f142))

## [1.0.3](https://github.com/silence-operator/silence-operator/compare/v1.0.2...v1.0.3) (2025-07-23)


### Bug Fixes

* **ci:** Disable CGO ([965a42c](https://github.com/silence-operator/silence-operator/commit/965a42cc2a0bbfefa0f5b50871d0df45164ae9ce))

## [1.0.2](https://github.com/silence-operator/silence-operator/compare/v1.0.1...v1.0.2) (2025-07-23)


### Bug Fixes

* **ci:** Use default baseimage for ko ([4b489c7](https://github.com/silence-operator/silence-operator/commit/4b489c7916f13920d497c299e0e724a6168968d0))
* **ci:** User personal token ([db23255](https://github.com/silence-operator/silence-operator/commit/db232555aa2a1bdaae3fce31d18770a690080f37))

## [1.0.1](https://github.com/silence-operator/silence-operator/compare/v1.0.0...v1.0.1) (2025-07-23)


### Bug Fixes

* **ci:** Update goreleaser workflow ([dd518d1](https://github.com/silence-operator/silence-operator/commit/dd518d119c79a7f925f67c82dd089616037edfd7))

## 1.0.0 (2025-07-23)


### Features

* **ci:** Add workflows ([b26e2f8](https://github.com/silence-operator/silence-operator/commit/b26e2f8157861a1d74d5f379ceaaddc1f7f6e5c4))
* **ci:** Add workflows ([cde273e](https://github.com/silence-operator/silence-operator/commit/cde273ee85844339f645b6d8005fb7271051c8bf))
* **deps:** Upgrade project with the latest kubebuilder version ([e479a9f](https://github.com/silence-operator/silence-operator/commit/e479a9f0ae35885753130abb2b0a560a971150f9))
* **deps:** Upgrade project with the latest kubebuilder version ([225e693](https://github.com/silence-operator/silence-operator/commit/225e69313aa965b58c40a85deb9cbd5ad60276b8))
* **helm:** Add servicemonitor ([7b2df22](https://github.com/silence-operator/silence-operator/commit/7b2df22ab4cd2ca12e7d8e5c0858e4129b9078ef))
* **helm:** Add servicemonitor ([3addfec](https://github.com/silence-operator/silence-operator/commit/3addfec346a863e38cea2d8a50beab505adb349e))
* **helm:** Update helm chart ([1d9dc35](https://github.com/silence-operator/silence-operator/commit/1d9dc358160ffe917202bb8af0281379b33b1f49))
* **helm:** Update helm chart ([7c06b60](https://github.com/silence-operator/silence-operator/commit/7c06b60a3239c41b7b3ec7d2ac15e5e9287a7b49))


### Bug Fixes

* **ci:** Add permissions to release please ([293ccdd](https://github.com/silence-operator/silence-operator/commit/293ccdd55af960dc252ce8666b8d6cd21380838b))
* **ci:** Disable PR trigger for goreleaser ([39c551d](https://github.com/silence-operator/silence-operator/commit/39c551da8cafa2bfb0803c04adcd5aa00625c58b))
* **ci:** Run linter on PR only ([1852862](https://github.com/silence-operator/silence-operator/commit/1852862047c5887d4178e07880fc6f0a7875b691))
* **ci:** Update goreleaser configuration ([b726f7a](https://github.com/silence-operator/silence-operator/commit/b726f7a155dc3aa28b1ac42c4fff0c48637af122))
