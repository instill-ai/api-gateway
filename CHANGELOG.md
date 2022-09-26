# Changelog

## [2.4.0](https://www.github.com/instill-ai/api-gateway/compare/v2.3.1...v2.4.0) (2022-02-28)


### Features

* **pipeline-backend:** add support URL/base64 endpoint and human-readable name ([#112](https://www.github.com/instill-ai/api-gateway/issues/112)) ([a6b9468](https://www.github.com/instill-ai/api-gateway/commit/a6b946838510853842993c2b3c269766c4b79ae5))


### Miscellaneous

* add project management workflows ([#113](https://www.github.com/instill-ai/api-gateway/issues/113)) ([1c600c4](https://www.github.com/instill-ai/api-gateway/commit/1c600c40f72ee9a57a7ee7ade698b3d4c751578c))

### [2.3.1](https://www.github.com/instill-ai/api-gateway/compare/v2.3.0...v2.3.1) (2022-01-28)


### Bug Fixes

* [pipeline-backend] add missing files and change file reference ([4540e03](https://www.github.com/instill-ai/api-gateway/commit/4540e036f558681359dedf8645a6f140b6624a81))
* [pipeline-backend] add more test cases ([07a2178](https://www.github.com/instill-ai/api-gateway/commit/07a2178852e7203c76397b10ad75026ea7bf4046))
* [pipeline-backend] remove useless test case for cubo ([645d58a](https://www.github.com/instill-ai/api-gateway/commit/645d58a000fbb183a4976aa47585cff2b3b1f933))
* add partial grpc e2e test and bug fixing ([09cf52a](https://www.github.com/instill-ai/api-gateway/commit/09cf52a5889095b1f3622f6f2ac00c535b903924))
* remove printing time measurement code ([d2958ad](https://www.github.com/instill-ai/api-gateway/commit/d2958ad6e256e8ae445878fe38fae11c42f19e5a))

## [2.3.0](https://www.github.com/instill-ai/api-gateway/compare/v2.2.0...v2.3.0) (2022-01-26)


### Features

* add gRPC and gRPC-gateway support ([8338096](https://www.github.com/instill-ai/api-gateway/commit/833809673e62ed3b374cf6ceb5e08710da3370f2))


### Bug Fixes

* remove code for debugging and future use purpose ([6927226](https://www.github.com/instill-ai/api-gateway/commit/69272260e5d9b64569f355a2eea6a75c8176d737))

## [2.2.0](https://www.github.com/instill-ai/api-gateway/compare/v2.1.0...v2.2.0) (2022-01-25)


### Features

* [kratos] add username in identity schema ([e29f4fb](https://www.github.com/instill-ai/api-gateway/commit/e29f4fb3ba503a3314da79f58ac074d990dd5672))
* add e2e test for logo classification ([ac8bde8](https://www.github.com/instill-ai/api-gateway/commit/ac8bde82cc9fd9f5fe2d17d83ec649c9589397a7))


### Bug Fixes

* remove cubo e2e test ([b4a83ef](https://www.github.com/instill-ai/api-gateway/commit/b4a83efe7dcebaecbb1bb873bddb26c60a5d90f3))
* remove plugin access ([ebf4efc](https://www.github.com/instill-ai/api-gateway/commit/ebf4efcbdac65cd86b06835a449b50b9fdf1429d))
* update inference-backend version which add classification by model endpoint ([2205ce5](https://www.github.com/instill-ai/api-gateway/commit/2205ce59d889d15ce732d16837aad3b0178a60fc))
* update model-backend version ([4cb1166](https://www.github.com/instill-ai/api-gateway/commit/4cb1166bf493e9ad5116ded7c70e0966e15a48d1))

## [2.1.0](https://www.github.com/instill-ai/api-gateway/compare/v2.0.5...v2.1.0) (2021-12-29)


### Features

* add `Backend` header in response for logging ([32befeb](https://www.github.com/instill-ai/api-gateway/commit/32befeb2594bd485b708e4e46602c68fb24cc38f))
* add json logging for Cloud logging ([caf4eb4](https://www.github.com/instill-ai/api-gateway/commit/caf4eb43225e624cbaf67c33c68977296d9e27bf))


### Bug Fixes

* add handling response with empty body ([038e484](https://www.github.com/instill-ai/api-gateway/commit/038e484b005d54c4c2c513de3236a4fda5ba0d6b))
* change pipeline to singular form ([da73a32](https://www.github.com/instill-ai/api-gateway/commit/da73a32ded3faf8aebaede4839096e7991c6f91b))
* disable JSON format logging for `__health` ([1fcf118](https://www.github.com/instill-ai/api-gateway/commit/1fcf118234f61b018326cde070b8555aaf793ec3))
* remove logging from client handler ([ee10325](https://www.github.com/instill-ai/api-gateway/commit/ee10325498d847c93e80c977d58d91fa74be04cb))


### Refactor

* change contact to support@instill.tech ([af9650e](https://www.github.com/instill-ai/api-gateway/commit/af9650eb7d06091509ae9e354b7fd649f3b1472a))
* change go-logging level to INFO ([edeac11](https://www.github.com/instill-ai/api-gateway/commit/edeac1115ee7d438328af94ae7b92c84f1ca7aa1))

### [2.0.5](https://www.github.com/instill-ai/api-gateway/compare/v2.0.4...v2.0.5) (2021-12-22)


### Bug Fixes

* fix model-backend dependency ([e170d99](https://www.github.com/instill-ai/api-gateway/commit/e170d9906ec93707fd883fc4ec819e039cf60ced))
* remove db-sql services ([9e1bc8f](https://www.github.com/instill-ai/api-gateway/commit/9e1bc8f1e4399166eae7ad414d8f559d1d20bdc9))


### Miscellaneous

* update mgmt-backend to 1.6.6 ([8cba865](https://www.github.com/instill-ai/api-gateway/commit/8cba865964a38d6cfd1694c952890d27117f17fc))

### [2.0.4](https://www.github.com/instill-ai/api-gateway/compare/v2.0.3...v2.0.4) (2021-12-22)


### Bug Fixes

* re-protect cubo endpoint ([2144ef0](https://www.github.com/instill-ai/api-gateway/commit/2144ef05290d6b09178608ec73cccdb33f2fc612))
* update tests ([442d9e3](https://www.github.com/instill-ai/api-gateway/commit/442d9e3fcc7421ea67b047a729fd8eea140f0bf6))


### Miscellaneous

* bump service version ([a470c3b](https://www.github.com/instill-ai/api-gateway/commit/a470c3bf95aff2a32996cb91778050e70af215d6))
* remove obsolete purge_harbor.sh ([27ec2db](https://www.github.com/instill-ai/api-gateway/commit/27ec2dbd8c0b2eef7aa4386ad007efd9aeaafbbb))

### [2.0.3](https://www.github.com/instill-ai/api-gateway/compare/v2.0.2...v2.0.3) (2021-12-11)


### Miscellaneous

* bump public version to 0.1.3 ([985a833](https://www.github.com/instill-ai/api-gateway/commit/985a833bd72433de9426935d215ec7908dbedb9c))

### [2.0.2](https://www.github.com/instill-ai/api-gateway/compare/v2.0.1...v2.0.2) (2021-12-11)


### Miscellaneous

* bump public version to 0.1.2 ([46eff15](https://www.github.com/instill-ai/api-gateway/commit/46eff151321f7e2b130a63fe46e614857ac7197c))

### [2.0.1](https://www.github.com/instill-ai/api-gateway/compare/v2.0.0...v2.0.1) (2021-12-09)


### Refactor

* print public version in script ([0fa0e0c](https://www.github.com/instill-ai/api-gateway/commit/0fa0e0cf749239e83678e917bbc75115213884d0))


### Miscellaneous

* remove creating public tag step ([09b001e](https://www.github.com/instill-ai/api-gateway/commit/09b001e80962ff546fbdd50d88ee43ff19c262c8))

## [2.0.0](https://www.github.com/instill-ai/api-gateway/compare/v1.16.1...v2.0.0) (2021-12-09)


### ⚠ BREAKING CHANGES

* [pipeline-backend] migrate from MySQL to PostgreSQL

### Bug Fixes

* [mgmt-backend] fix migration db config ([1a593db](https://www.github.com/instill-ai/api-gateway/commit/1a593dbdf4f87716f11507ff1213d032a1a13bd2))
* add log file ([dcf1b3c](https://www.github.com/instill-ai/api-gateway/commit/dcf1b3c57c1ee915a4c09a5c690a03a5c4e02d26))
* change model-backend env variables ([63c91c4](https://www.github.com/instill-ai/api-gateway/commit/63c91c4a5675acb3a509d07db8154222d396e62b))
* db-sql env ([1198e08](https://www.github.com/instill-ai/api-gateway/commit/1198e086c601077f9ad2cc06a1c84c7f4933d3d4))
* fix permission issue ([031bb43](https://www.github.com/instill-ai/api-gateway/commit/031bb43eeb6b80af28cdcdcb57894a1f6b5bedb9))
* update identity scheme ([352c3ca](https://www.github.com/instill-ai/api-gateway/commit/352c3cae19aac4b0110ee52414f6fa5ad252f5a1))


### Refactor

* [model-backend] Switch from mysql to postgres database ([46e2ed2](https://www.github.com/instill-ai/api-gateway/commit/46e2ed2ed25add1ff0829ada2725363a13c64ab4))
* [pipeline-backend] migrate from MySQL to PostgreSQL ([a60c4e7](https://www.github.com/instill-ai/api-gateway/commit/a60c4e7dae78cacf7ac0a5da8394f83429ab2162))
* use the same env style ([274e8ea](https://www.github.com/instill-ai/api-gateway/commit/274e8eaa94bd8d88aef56761d02f3d8f70211567))


### Miscellaneous

* [model-backend] bump up version to 2.0.1 ([4d1de37](https://www.github.com/instill-ai/api-gateway/commit/4d1de37176a30f42b5ebf2fb50e736b9ef4baa04))
* bump inference backend to 1.5.11, mgmt-backend to 1.6.3 ([5f82851](https://www.github.com/instill-ai/api-gateway/commit/5f8285122a991653ae497379f84289a813b9f0d9))
* bump inference, pipeline, mgmt backends ([5d0c5fe](https://www.github.com/instill-ai/api-gateway/commit/5d0c5fe6f233c75e5e8d3f00ba9e4ed4b597a029))
* decouple integration-test into steps ([03f2329](https://www.github.com/instill-ai/api-gateway/commit/03f232955b2804033cf8deb22d281af4e4333564))
* fix purge core version regex when release ([8832108](https://www.github.com/instill-ai/api-gateway/commit/883210888206c245b4e58a8a3dba4cb4ea08fea2))
* unify docker-compose service and container names ([91d0e70](https://www.github.com/instill-ai/api-gateway/commit/91d0e707f70729cdf0100c3182b1df285e3620c4))
* update model-backend to 2.0.3 ([533a6be](https://www.github.com/instill-ai/api-gateway/commit/533a6beaebaee0c1d1eb9417c682e7bbe0b51d1a))
* use a non-root postgresql image ([454c66b](https://www.github.com/instill-ai/api-gateway/commit/454c66b6049f14069db1d48da62a3d76496ae8bd))

### [1.16.1](https://www.github.com/instill-ai/api-gateway/compare/v1.16.0...v1.16.1) (2021-12-07)


### Bug Fixes

* [pipeline-backend] add test script ([11fc5c2](https://www.github.com/instill-ai/api-gateway/commit/11fc5c29d68ac7daecd216b2708db366a1282ad5))
* [pipeline-backend] fix TYPO in test script ([66ff759](https://www.github.com/instill-ai/api-gateway/commit/66ff759842c4b0d15d1d6fbcac0408206541a17b))


### Miscellaneous

* [inference-backend] bump up version to 1.5.10 ([c4820f8](https://www.github.com/instill-ai/api-gateway/commit/c4820f86cc228a008c3da719515188e29e62847d))

## [1.16.0](https://www.github.com/instill-ai/api-gateway/compare/v1.15.4...v1.16.0) (2021-12-06)


### Features

* [pipeline-backend] add query parameter for listing pipeline ([3a4c292](https://www.github.com/instill-ai/api-gateway/commit/3a4c292abde14ee2de14a26cf1fa95671faff8c0))


### Miscellaneous

* bump up public-version to 0.1.0 ([6eb49dd](https://www.github.com/instill-ai/api-gateway/commit/6eb49dd4466e9842b4a103de2cebbcd9e7a7d38c))

### [1.15.4](https://www.github.com/instill-ai/api-gateway/compare/v1.15.3...v1.15.4) (2021-12-05)


### Bug Fixes

* add temp test for temp no auth endpoint ([e48511b](https://www.github.com/instill-ai/api-gateway/commit/e48511ba149dc4022059a7806ba49a23b2a1b4a8))
* add timeout ([52439cf](https://www.github.com/instill-ai/api-gateway/commit/52439cfc8221acebc0a3a6af8d96d18d645ea8ec))
* make inference backend template consistent ([1717e8a](https://www.github.com/instill-ai/api-gateway/commit/1717e8a04226a462615d9e02e46372fb6133fc5f))
* rename auth type ([b15c5bb](https://www.github.com/instill-ai/api-gateway/commit/b15c5bbbfa34b941873c070d1bc7590d4732bf40))
* simplify headers to pass ([6ff32f3](https://www.github.com/instill-ai/api-gateway/commit/6ff32f3ced21591dd82dc9589fb695fcc6e96a53))
* temporally unprotect /tasks/detection/models/{model}/versions/{version}/outputs ([252c585](https://www.github.com/instill-ai/api-gateway/commit/252c5857c30481af2e7608be6a57bcc5099face9))


### Miscellaneous

* bump up mgmt-backend to 1.6.0 ([af58ff7](https://www.github.com/instill-ai/api-gateway/commit/af58ff7a217cb1991d7370885593bb10a82befb8))
* update openapi.yaml ([01eab96](https://www.github.com/instill-ai/api-gateway/commit/01eab96bbe9a8e2ff99bb268cb5a061220d9d9dc))


### Refactor

* rename pipeline-backend to singular form ([d945cd3](https://www.github.com/instill-ai/api-gateway/commit/d945cd36d6ddd10a79ea4ae3076aaffa28710454))

### [1.15.3](https://www.github.com/instill-ai/api-gateway/compare/v1.15.2...v1.15.3) (2021-11-30)


### Bug Fixes

* add null request body for GET operation ([9b36c87](https://www.github.com/instill-ai/api-gateway/commit/9b36c87b69f23c369da1be5bdc776fca12546ccb))
* check rounded up 'updated_at' ([674d901](https://www.github.com/instill-ai/api-gateway/commit/674d9019565150b9e23d089162fb7edfcc22e1f1))
* mount TLS to mgmt-backend-igrate ([adbcf85](https://www.github.com/instill-ai/api-gateway/commit/adbcf8514844a1ce19e999a747e80e8cddef2f92))
* refactor tests to only wipe out test data ([ce1dddd](https://www.github.com/instill-ai/api-gateway/commit/ce1dddd4628d30a38a0423139e8c0dc1ac00a36f))


### Miscellaneous

* bump up version to 1.5.7 ([9164858](https://www.github.com/instill-ai/api-gateway/commit/9164858f41a1f4e06e2ed1ccdf743c4e746f5126))
* update openapi.yaml ([2a6e3b1](https://www.github.com/instill-ai/api-gateway/commit/2a6e3b1180c2953fd8baa0e2010c13ebb36635b7))

### [1.15.2](https://www.github.com/instill-ai/api-gateway/compare/v1.15.1...v1.15.2) (2021-11-28)


### Bug Fixes

* rename 'ISSUER' to 'HYDRA_ISSUER' ([a7901de](https://www.github.com/instill-ai/api-gateway/commit/a7901dec871d263f733664417f644a401f7e3d9f))

### [1.15.1](https://www.github.com/instill-ai/api-gateway/compare/v1.15.0...v1.15.1) (2021-11-26)


### Bug Fixes

* add description check ([69dafc5](https://www.github.com/instill-ai/api-gateway/commit/69dafc5192f87a785e47acc806b0e7eba31ba183))
* update mgmt-backend ([473480b](https://www.github.com/instill-ai/api-gateway/commit/473480b2784344aa94b23e58e43004a341f7fe2b))
* update tests ([19e5aff](https://www.github.com/instill-ai/api-gateway/commit/19e5affa623f7da91c21494de288e97bde760c3f))


### Miscellaneous

* bump up mgmt-backend to 1.5.4 ([4103410](https://www.github.com/instill-ai/api-gateway/commit/4103410e6d36a602086a7ada79f197a4ee374d83))

## [1.15.0](https://www.github.com/instill-ai/api-gateway/compare/v1.14.0...v1.15.0) (2021-11-23)


### Features

* update model backend with model description ([a5a6d41](https://www.github.com/instill-ai/api-gateway/commit/a5a6d4140059eb3419f01d69c34652d9c6d35d81))


### Miscellaneous

* fix purge core version regex ([f0a3b25](https://www.github.com/instill-ai/api-gateway/commit/f0a3b2509153a6bea82b8dd84ca611032c139f82))

## [1.14.0](https://www.github.com/instill-ai/api-gateway/compare/v1.13.1...v1.14.0) (2021-11-23)


### Features

* publish api-gateway openapi.yaml (merge all backend openapi.yaml files) ([3cba7ea](https://www.github.com/instill-ai/api-gateway/commit/3cba7ea09fb4b0c9528bf747b189e9e6b2eeeaa1))

### [1.13.1](https://www.github.com/instill-ai/api-gateway/compare/v1.13.0...v1.13.1) (2021-11-19)


### Bug Fixes

* fix api testing spec ([8bf5a06](https://www.github.com/instill-ai/api-gateway/commit/8bf5a06ad90136d51660e9252fb3c33d8472e41d))
* fix duplicate keys ([dd0176f](https://www.github.com/instill-ai/api-gateway/commit/dd0176fca2c8fc73a8ba617ddd8b64b32cb1dc3a))
* rare condition in model get latest api by add sleep before check function ([766bb8a](https://www.github.com/instill-ai/api-gateway/commit/766bb8a7b63cfe3e26bfa4c2d981e0d6ef566bbb))
* update model field name from name_id to ext_id in model backend test cases ([51b5338](https://www.github.com/instill-ai/api-gateway/commit/51b53382332cafd8c0fdd30256d3cffd1cad282c))


### Refactor

* move all backend configs into backends folder ([ef141bb](https://www.github.com/instill-ai/api-gateway/commit/ef141bb687875f34fc8af165135b422bd281c3d9))


### Miscellaneous

* [inference-backend] bump up to v1.5.9 ([1e6e824](https://www.github.com/instill-ai/api-gateway/commit/1e6e824223531e4dfa5552c86cf6589757d17aa8))
* [inference-backend] remove model config ([bbc3535](https://www.github.com/instill-ai/api-gateway/commit/bbc35351da58a77e9e233c5730d45044f35b420e))
* fix cd purge logic ([d05181f](https://www.github.com/instill-ai/api-gateway/commit/d05181ff2d36b20f193c4d89c13e335d5fa6e5fc))

## [1.13.0](https://www.github.com/instill-ai/api-gateway/compare/v1.12.5...v1.13.0) (2021-11-12)


### Features

* add hydra in local dev to issue user token ([cf44cbf](https://www.github.com/instill-ai/api-gateway/commit/cf44cbfd318d1caee2a8348a0ac01edcdf5c2743))
* add k6 test for hydra issued user token ([6d8a5a3](https://www.github.com/instill-ai/api-gateway/commit/6d8a5a32ed061d69154a04afc5929544a3a52d10))
* add k6 test for krato to create test user ([0558840](https://www.github.com/instill-ai/api-gateway/commit/055884099f0f5a325f669d03265767a80b6f6b0a))
* add kratos ([7f88106](https://www.github.com/instill-ai/api-gateway/commit/7f88106f2ea681f1234b7cfa73a43c81656475d4))
* add model backend ([6057625](https://www.github.com/instill-ai/api-gateway/commit/6057625a904f4abf4ae2fa5f0c9fc3f1cb88fcb3))
* expose pipeline-backend endpoints ([cf898a2](https://www.github.com/instill-ai/api-gateway/commit/cf898a287687484fb02ce5a1ec468f5c4e452dc9))
* gen certs for hydra ([1180bda](https://www.github.com/instill-ai/api-gateway/commit/1180bdaf24bbf4512ca435dace89d5ffdd00d4da))
* gen certs for kratos ([c7fe5a0](https://www.github.com/instill-ai/api-gateway/commit/c7fe5a00daf157f91d815bb0eb0e09406ab91b0f))


### Bug Fixes

* add complete inference round trip ([9f9a922](https://www.github.com/instill-ai/api-gateway/commit/9f9a92298191ac7dfc42bb667a5e0fb0cee12264))
* add hydra ([e59c3df](https://www.github.com/instill-ai/api-gateway/commit/e59c3df77db7806bb4c55816409bb22d12ba5c56))
* add hydra port env configuration ([645f35f](https://www.github.com/instill-ai/api-gateway/commit/645f35ffb3c818f5e6653d99aa9147210819cc67))
* make aud and iss configurable ([aedd35f](https://www.github.com/instill-ai/api-gateway/commit/aedd35f87e6ed8293453b74e039129eb3763fd1a))
* move test code to another file ([9a4fc2c](https://www.github.com/instill-ai/api-gateway/commit/9a4fc2c1d7afd6e2f4f03e796ccbc5ef8b39d97e))
* refactor access token payload ([96d5053](https://www.github.com/instill-ai/api-gateway/commit/96d5053dbf6ced0553b2e3c2244174faa5a626f8))
* replace auth0 with hydra ([4329f93](https://www.github.com/instill-ai/api-gateway/commit/4329f937903261b428b719b8ecd077ce7bac311d))
* use env specific iss and aud ([ef1c64a](https://www.github.com/instill-ai/api-gateway/commit/ef1c64a46e110d0248fc1b3256f3b7ea8d340dc1))
* use fixed issuer ([bd645e3](https://www.github.com/instill-ai/api-gateway/commit/bd645e30aaee56e265bfb7b4f4130246a7d4ef6c))
* use hydra jwk and remove scopes ([cd224de](https://www.github.com/instill-ai/api-gateway/commit/cd224de8b79f5b3d581ee934561c163d330bc4f1))
* use HYDRA_ prefix to make the config clear ([853586e](https://www.github.com/instill-ai/api-gateway/commit/853586e212af4584e12e73a32353ca7353535fed))


### Refactor

* make all endpoints use the same issuer ([c7e3d62](https://www.github.com/instill-ai/api-gateway/commit/c7e3d62fdd772f6c96cb1920681cb6b217398b4c))
* output golang version ([501f1e1](https://www.github.com/instill-ai/api-gateway/commit/501f1e1476334d0a9317da9f6c81149b273a301d))
* rename port name ([6e518da](https://www.github.com/instill-ai/api-gateway/commit/6e518daf979affa41928055950d821cf04c74535))
* use internal dns ([6b4cd42](https://www.github.com/instill-ai/api-gateway/commit/6b4cd42db42af61c016ee70ceb2bec1b1e61aed4))


### Miscellaneous

* add all new services in Makefile ([9c24ec7](https://www.github.com/instill-ai/api-gateway/commit/9c24ec785d83e83a7c840178ecbf5ce7113de0b2))
* add CODEOWNERS ([5359947](https://www.github.com/instill-ai/api-gateway/commit/53599477669f75b60ed039ee82371bd46984977b))
* bump up inference-backend to 1.5.7 ([f5939d5](https://www.github.com/instill-ai/api-gateway/commit/f5939d5c4848af787d272a38672aa272e06e19e7))
* bump up inference-backend to 1.5.8 ([ed1c2f7](https://www.github.com/instill-ai/api-gateway/commit/ed1c2f7000f53a13ae07ae671494c655db36bab7))
* setup gh actions with specific go version ([62ed12c](https://www.github.com/instill-ai/api-gateway/commit/62ed12c5b9bcd9c45902caf53c77ebf0535cfd1b))
* update .pre-commit-config.yaml ([d09fc66](https://www.github.com/instill-ai/api-gateway/commit/d09fc66cd49f54cd1849ab9fb12eefcb36976b86))
* update cd script ([0f545b0](https://www.github.com/instill-ai/api-gateway/commit/0f545b07d6099bddd1e4073c64e10255b59c87a8))
* update ci script ([2452cc0](https://www.github.com/instill-ai/api-gateway/commit/2452cc053c1b8bc161ddda8e22b96d99b3c3a0dc))
* update ci/cd script ([b7912fd](https://www.github.com/instill-ai/api-gateway/commit/b7912fdc2a29c6c4334e586d59ed26e902e5a0ab))
* update PR template ([63d08fc](https://www.github.com/instill-ai/api-gateway/commit/63d08fc44dc8059ee34b1a737bedd919376efa88))
* upgrade krakend to 1.4.1 ([8df6b67](https://www.github.com/instill-ai/api-gateway/commit/8df6b67705a4c1079541d5d009c49c2b9146ddad))

### [1.12.5](https://www.github.com/instill-ai/api-gateway/compare/v1.12.4...v1.12.5) (2021-10-11)


### Bug Fixes

* add Request-Id in http handler ([b6d173e](https://www.github.com/instill-ai/api-gateway/commit/b6d173e89c3f50a0a8e895dca136034c51ff9f92))
* pass Request-Id to all backends ([78b9d13](https://www.github.com/instill-ai/api-gateway/commit/78b9d1395ddbd1235bedc7766dabc3058673aaff))


### Miscellaneous

* bump up mgmt-backend to 1.5.1 ([e476f7f](https://www.github.com/instill-ai/api-gateway/commit/e476f7fe5596c48f44c16661fbafe7e57aa8027e))

### [1.12.4](https://www.github.com/instill-ai/api-gateway/compare/v1.12.3...v1.12.4) (2021-09-23)


### Refactor

* simplify docker images ([b77d247](https://www.github.com/instill-ai/api-gateway/commit/b77d247935cb7ada2ed574c2d14ff2b1c93983c6))


### Miscellaneous

* fix ci/cd script ([41102ac](https://www.github.com/instill-ai/api-gateway/commit/41102ac91d09791e9159a0ee41882aa09549858f))
* update release changelog-types ([058742d](https://www.github.com/instill-ai/api-gateway/commit/058742dca9f022054d3971014ffee5d555af16d3))
* upgrade pre-commit-hooks version ([e9a3ce5](https://www.github.com/instill-ai/api-gateway/commit/e9a3ce59a4b105fbe62d036fe2d1bbade4418e22))

### [1.12.3](https://www.github.com/instill-ai/api-gateway/compare/v1.12.2...v1.12.3) (2021-09-04)


### Bug Fixes

* fix demo endpoints ([5dc54ef](https://www.github.com/instill-ai/api-gateway/commit/5dc54ef6af5eafa3bd49743d5f8fbded87df624c))

### [1.12.2](https://www.github.com/instill-ai/api-gateway/compare/v1.12.1...v1.12.2) (2021-09-04)


### Bug Fixes

* fix api-gateway dockerfile ([74ea1a0](https://www.github.com/instill-ai/api-gateway/commit/74ea1a0584aba15538f0e3a69cf0a5c0d3ddd8d4))

### [1.12.1](https://www.github.com/instill-ai/api-gateway/compare/v1.12.0...v1.12.1) (2021-08-28)


### Miscellaneous

* fix typos in tests ([060d8af](https://www.github.com/instill-ai/api-gateway/commit/060d8af36e0f92d38b336df8f4c1801b89efa4fa))
* update api-gateway Dockerfile ([d66b88e](https://www.github.com/instill-ai/api-gateway/commit/d66b88e3d5116cd9d23c1cd6a8e64463d2d5e9cc))

## [1.12.0](https://www.github.com/instill-ai/api-gateway/compare/v1.11.0...v1.12.0) (2021-08-04)


### Features

* add inference api health check endpoint ([0422f7a](https://www.github.com/instill-ai/api-gateway/commit/0422f7ac4f0845c18cff1ca35a45be7c4a812cc8))


### Miscellaneous

* bump up elastic stack version to 7.13.2 ([2d184d7](https://www.github.com/instill-ai/api-gateway/commit/2d184d7ea4e750cf0bbed93d978e390a262b0590))
* bump up inference-backend version to 1.4.0 ([b5162b8](https://www.github.com/instill-ai/api-gateway/commit/b5162b8144caf1f514f972ad18ac3764fd3d611e))
* bump up inference-backend version to 1.4.1 ([bd4bdfe](https://www.github.com/instill-ai/api-gateway/commit/bd4bdfee7e24f05ebc6caa66ac55c3129481b28c))
* improve local-check-script.js ([e0a1ebd](https://www.github.com/instill-ai/api-gateway/commit/e0a1ebd7e61a80bbdc3fe4160ac54749ff81402b))
* make endpoint method and timeout configurable in api-gateway config ([1fea04b](https://www.github.com/instill-ai/api-gateway/commit/1fea04b47a234cc8231bc237b813590f4a2c7a1e))
* rename test scripts ([29b15a8](https://www.github.com/instill-ai/api-gateway/commit/29b15a883898c618a963613ae1b654dba3dc69f1))
* update inference-backend config file ([635207e](https://www.github.com/instill-ai/api-gateway/commit/635207e41af899b3b93f5cd920c0b4f1fee26930))
* update krakend config files ([a2e7769](https://www.github.com/instill-ai/api-gateway/commit/a2e7769cca91a73f64547d8fdd4aeb7542d8fa72))

## [1.11.0](https://www.github.com/instill-ai/api-gateway/compare/v1.10.1...v1.11.0) (2021-06-11)


### Features

* add integration tests ([4450e57](https://www.github.com/instill-ai/api-gateway/commit/4450e5710ff3768d41f43364d89e5c5d936ebe40))


### Bug Fixes

* enrich log and error message ([0d38eb6](https://www.github.com/instill-ai/api-gateway/commit/0d38eb64f5d9a6d9aedba4a7ebc310a6195b0b83))
* use db-sql dev image ([89ef654](https://www.github.com/instill-ai/api-gateway/commit/89ef6541bd7530f7dc9c79ba82fabc3d5c42572f))


### Miscellaneous

* add cubo tests ([129da4b](https://www.github.com/instill-ai/api-gateway/commit/129da4bdcff70da3d231bfb333a2791a7e53380c))
* add inference api tests ([e3152ca](https://www.github.com/instill-ai/api-gateway/commit/e3152ca2ea7503b14719694a540f3439a29ebc03))
* bump up lolocal mgmt-backend and db version ([2097d63](https://www.github.com/instill-ai/api-gateway/commit/2097d63a824975c1d58bf443fd7ff9d5075e82f1))
* clean up env variables ([a46a286](https://www.github.com/instill-ai/api-gateway/commit/a46a286c5aa3af82f6178e997dd85b06ffc5ee54))
* fix k6 script ([a21efc8](https://www.github.com/instill-ai/api-gateway/commit/a21efc8e2dae5777b15de24c93604f455571e7ac))
* lock influxdb at version 1.8 ([c0c1987](https://www.github.com/instill-ai/api-gateway/commit/c0c198776aa5b96ccd0c44c77fb22e43a8aa398a))
* remove error 400 response detail overwriting logic ([02fe25d](https://www.github.com/instill-ai/api-gateway/commit/02fe25d0b143c5870633dd5d9dd8904ecc0a5798))
* remove root user in docker-compose ([9358e41](https://www.github.com/instill-ai/api-gateway/commit/9358e413bb8f77a86c8cbb59fc3d6bc56305e6c9))
* remove useless .env variables ([6d54852](https://www.github.com/instill-ai/api-gateway/commit/6d54852d80d561d28cbe8ce7a1d67b17927bedf5))
* update config files ([e8bf964](https://www.github.com/instill-ai/api-gateway/commit/e8bf9644591fdbfff40075f5ff92262dfb4eac92))
* update modifier plug-in ([7c230f6](https://www.github.com/instill-ai/api-gateway/commit/7c230f6e660410ca623034e1375e4022d7ce60cd))
* update service version ([b4ca0c6](https://www.github.com/instill-ai/api-gateway/commit/b4ca0c6883c0b8dfc61a36fd85ad4aeab5697985))

### [1.10.1](https://www.github.com/instill-ai/api-gateway/compare/v1.10.0...v1.10.1) (2021-05-17)


### Bug Fixes

* fix scope for detection endpoint ([464a461](https://www.github.com/instill-ai/api-gateway/commit/464a46178f43bad989fe1b0560e5742f19852879))


### Miscellaneous

* update all config files ([b8c66b6](https://www.github.com/instill-ai/api-gateway/commit/b8c66b63d9db3dd2030c43a33a724a490667b60e))

## [1.10.0](https://www.github.com/instill-ai/api-gateway/compare/v1.9.5...v1.10.0) (2021-05-17)


### Features

* make duration field inserted in handler plugin ([4ef3c8c](https://www.github.com/instill-ai/api-gateway/commit/4ef3c8cfff94511ca2200153e127ebf50004d951))


### Miscellaneous

* bump up local inference-backend and mgmt-backend image version ([91459d5](https://www.github.com/instill-ai/api-gateway/commit/91459d54342d6932e9a7e1df88e4a7f7256cae5a))

### [1.9.5](https://www.github.com/instill-ai/api-gateway/compare/v1.9.4...v1.9.5) (2021-05-16)


### Miscellaneous

* add created_ts and duration in response json ([cb67a34](https://www.github.com/instill-ai/api-gateway/commit/cb67a34ac9b404839e25eb6ec7090c12d3001a20))
* add error check for ioutil.ReadAll ([76941a2](https://www.github.com/instill-ai/api-gateway/commit/76941a277a5ea23015ddc4f0bf984a7ab9711d65))

### [1.9.4](https://www.github.com/instill-ai/api-gateway/compare/v1.9.3...v1.9.4) (2021-05-13)


### Miscellaneous

* add 404 and 405 response error json ([9310e85](https://www.github.com/instill-ai/api-gateway/commit/9310e85bf392b50cd22fcc57f4b96e139ab203c8))

### [1.9.3](https://www.github.com/instill-ai/api-gateway/compare/v1.9.2...v1.9.3) (2021-05-12)


### Bug Fixes

* downgrade golang base image to 1.15.8 (need to match krakend 1.3.0) ([14578b1](https://www.github.com/instill-ai/api-gateway/commit/14578b142a9f44c3734f3eb7ad82411b33353c04))


### Miscellaneous

* upgrade to go 1.16 and enable module mode by default ([d2db44f](https://www.github.com/instill-ai/api-gateway/commit/d2db44f508a777cb6c7f25bd9511e190877f1818))

### [1.9.2](https://www.github.com/instill-ai/api-gateway/compare/v1.9.1...v1.9.2) (2021-05-12)


### Miscellaneous

* add JWT models validation logic ([ccc86bc](https://www.github.com/instill-ai/api-gateway/commit/ccc86bc0f9e275ce7b3545576ad10008758de783))
* remove plugin go module (not used) ([2a1749a](https://www.github.com/instill-ai/api-gateway/commit/2a1749ae0af92600ab1457208188680fa8044c28))

### [1.9.1](https://www.github.com/instill-ai/api-gateway/compare/v1.9.0...v1.9.1) (2021-05-12)


### Bug Fixes

* refactor backends ([0608ddb](https://www.github.com/instill-ai/api-gateway/commit/0608ddbad1223fb447a2bd94995fee90c2397aac))


### Miscellaneous

* bump up mgmt backend version to 1.3.3 ([05b56e2](https://www.github.com/instill-ai/api-gateway/commit/05b56e21ce4c3a8db28c3d6e4de7221446dabd54))

## [1.9.0](https://www.github.com/instill-ai/api-gateway/compare/v1.8.5...v1.9.0) (2021-05-09)


### Features

* add model-specific inference endpoints ([99da527](https://www.github.com/instill-ai/api-gateway/commit/99da52767e199481ff1b316cefa87c75b8f47130))


### Bug Fixes

* fix logging plugin returning wrong status code ([ed7ff9c](https://www.github.com/instill-ai/api-gateway/commit/ed7ff9c7a2e5a94843c11971cd77ad576479b5d4))


### Miscellaneous

* add post method for default endpoint ([ab2e0ee](https://www.github.com/instill-ai/api-gateway/commit/ab2e0ee32b2817227ef87a7ad8765534db821207))
* update api-gateway config files ([53d4597](https://www.github.com/instill-ai/api-gateway/commit/53d45975c6d541cba45a424b94d59dc23cae0e37))

### [1.8.5](https://www.github.com/instill-ai/api-gateway/compare/v1.8.4...v1.8.5) (2021-04-21)


### Miscellaneous

* fix krakend-metrics listen_address ([a290de2](https://www.github.com/instill-ai/api-gateway/commit/a290de28bd17a90843227e565e2077a8a1e43eb1))

### [1.8.4](https://www.github.com/instill-ai/api-gateway/compare/v1.8.3...v1.8.4) (2021-04-20)


### Miscellaneous

* enable all krakend-metrics ([a2bffdd](https://www.github.com/instill-ai/api-gateway/commit/a2bffdd2a919209fece2a02f8d494a175a74a737))

### [1.8.3](https://www.github.com/instill-ai/api-gateway/compare/v1.8.2...v1.8.3) (2021-04-20)


### Bug Fixes

* use auth0 custom domain for production ([91e7587](https://www.github.com/instill-ai/api-gateway/commit/91e75878056896e4395d85be7c02ab0eabcf5e07))

### [1.8.2](https://www.github.com/instill-ai/api-gateway/compare/v1.8.1...v1.8.2) (2021-04-18)


### Miscellaneous

* add OPTIONS in cors allow_methods ([0b625f9](https://www.github.com/instill-ai/api-gateway/commit/0b625f9fcee64af6f5e6a4f6636d0173feb4cd93))

### [1.8.1](https://www.github.com/instill-ai/api-gateway/compare/v1.8.0...v1.8.1) (2021-04-13)


### Miscellaneous

* bump up local dev backend versions ([9854764](https://www.github.com/instill-ai/api-gateway/commit/9854764e16801f02ef54aa23364282fdddc28e78))
* correct self url at api gateway using handler plugin ([ee340bb](https://www.github.com/instill-ai/api-gateway/commit/ee340bb83cc10939f6d66b3af15dd59552f49fe1))

## [1.8.0](https://www.github.com/instill-ai/api-gateway/compare/v1.7.2...v1.8.0) (2021-04-07)


### Features

* add demo endpoints for inference tasks ([e4237ce](https://www.github.com/instill-ai/api-gateway/commit/e4237ce7a371378f1b96b285fe27910621bef576))


### Bug Fixes

* make x-krakend-completed true with json outputs ([1d188c3](https://www.github.com/instill-ai/api-gateway/commit/1d188c3ac2851c9f51a67155429fc099c431938e))
* remove undefined logging for auth backend ([0fb4ff2](https://www.github.com/instill-ai/api-gateway/commit/0fb4ff2fbd9a6d219a8422599aaf3e4672693c78))


### Miscellaneous

* add jose scope validation for mgmt API ([8e25a5c](https://www.github.com/instill-ai/api-gateway/commit/8e25a5cd3ffc30d8852b3c66e1766982a229b9db))
* add jwt-aud header propagation for inference endpoints ([0233367](https://www.github.com/instill-ai/api-gateway/commit/0233367ae3770b6c40698ead72092b24c8403694))
* add krakend-cel for checking melicious jwt header hijacking from client ([249597b](https://www.github.com/instill-ai/api-gateway/commit/249597b792771e63fecf67bfdd9613bef842e6e9))
* add make config target for gereating all env configs ([7731208](https://www.github.com/instill-ai/api-gateway/commit/7731208e3f6f0b221b7ecd2845998ec7be518603))
* bump up auth-backend version for dev env ([4004ccc](https://www.github.com/instill-ai/api-gateway/commit/4004ccccc71352b94b5f98a5c428312f90be67a0))
* consolidate jwt headers ([4c70f38](https://www.github.com/instill-ai/api-gateway/commit/4c70f381049c5393c0391337762d155a1a9d8a89))
* fix service dependency and typos ([7379701](https://www.github.com/instill-ai/api-gateway/commit/73797014fe663fe198865ec5d123c838ced19286))
* remove api-gateway/config/krakend.json ([ab0087d](https://www.github.com/instill-ai/api-gateway/commit/ab0087dff7392f3ea8c7db8af323e59d8978e0ef))
* remove jwt-sub header propagation for inference endpoints ([31eb2a9](https://www.github.com/instill-ai/api-gateway/commit/31eb2a90887df417c39ab6877340aa6af1a6f5f3))
* resume no-op for forward-proxy endpoints ([49884b6](https://www.github.com/instill-ai/api-gateway/commit/49884b65909f57a2b49b7b549468f764f28d4ac3))
* unify jwt issuer env variables ([e23411b](https://www.github.com/instill-ai/api-gateway/commit/e23411bf7fca87ca3ee1a308921d9084f17b8d78))
* update krakend configuration files ([fe1af1c](https://www.github.com/instill-ai/api-gateway/commit/fe1af1c1130171cac0b0abe94639e12d6d835464))
* update krakend configuration files ([c61c5d9](https://www.github.com/instill-ai/api-gateway/commit/c61c5d91d19dddbc2f65577b59763b5233165757))
* update krakend configuration json files ([5b91ec7](https://www.github.com/instill-ai/api-gateway/commit/5b91ec756d4c3cb7a4682dbe69fad5d219c9f0bc))
* upgrade auth backend version to 1.1.5 ([8f278bb](https://www.github.com/instill-ai/api-gateway/commit/8f278bb7a503ea4e0fa957f9dd365df026a8f78f))

### [1.7.2](https://www.github.com/instill-ai/api-gateway/compare/v1.7.1...v1.7.2) (2021-04-03)


### Bug Fixes

* add make config-g0 and config-staging ([95b1418](https://www.github.com/instill-ai/api-gateway/commit/95b1418e931c95ca7c5ef847205ae5f85eda7ae3))
* build image for g0 and staging ([ef8f135](https://www.github.com/instill-ai/api-gateway/commit/ef8f135ec420674bd59a9cfe2bbeeb51337139a9))
* make auth0 validator dep of Auth0 domain ([a7ee599](https://www.github.com/instill-ai/api-gateway/commit/a7ee5998eb79fa49935034729dd0dddc484df77a))

### [1.7.1](https://www.github.com/instill-ai/api-gateway/compare/v1.7.0...v1.7.1) (2021-04-02)


### Bug Fixes

* add /users endpoints ([0556ece](https://www.github.com/instill-ai/api-gateway/commit/0556ece6e87bc6c46a2ec15be2bcd89b165dd710))


### Miscellaneous

* re-generate configuration files ([0580291](https://www.github.com/instill-ai/api-gateway/commit/05802914da6421a41044393ef31e6789543bb0e7))

## [1.7.0](https://www.github.com/instill-ai/api-gateway/compare/v1.6.1...v1.7.0) (2021-04-02)


### Features

* enable CORS configuration ([53ac3e4](https://www.github.com/instill-ai/api-gateway/commit/53ac3e416942175733561b57d95d652a0dbfda2a))


### Bug Fixes

* add CORS configuration ([7ebb373](https://www.github.com/instill-ai/api-gateway/commit/7ebb3736aee8448ed75ff4e09f92c5d30f877380))


### Miscellaneous

* upgrade AUTH_BACKEND_VERSION to 1.1.2 ([17fc6bc](https://www.github.com/instill-ai/api-gateway/commit/17fc6bc449b08f4c2b244e95bafb66675ce06e46))

### [1.6.1](https://www.github.com/instill-ai/api-gateway/compare/v1.6.0...v1.6.1) (2021-04-01)


### Bug Fixes

* update auth env variables ([d974054](https://www.github.com/instill-ai/api-gateway/commit/d974054eb61d39930fbc5db0435fd3feb0451ae1))
* update endpoint names ([8417f08](https://www.github.com/instill-ai/api-gateway/commit/8417f080aaa70dcd866649b1ffab8f21952bde88))


### Miscellaneous

* update krakend configuration ([4392d7b](https://www.github.com/instill-ai/api-gateway/commit/4392d7b16c783652df179cce28383ccd3dd228b6))
* update krakend dev configuration ([eba74c1](https://www.github.com/instill-ai/api-gateway/commit/eba74c122ab7a0e7ee18043ea773232e301c9163))

## [1.6.0](https://www.github.com/instill-ai/api-gateway/compare/v1.5.3...v1.6.0) (2021-03-31)


### Features

* add auth backend configuration ([d534d73](https://www.github.com/instill-ai/api-gateway/commit/d534d73174684c7ba4189f653baabe66451c05dd))
* add auth backend for local deployment ([01eb3fd](https://www.github.com/instill-ai/api-gateway/commit/01eb3fd02de06713c7728c30e472ae2906858785))
* add auth backend service ([b1a250a](https://www.github.com/instill-ai/api-gateway/commit/b1a250a902fdddc9cdce73151725286c25e60ea8))


### Bug Fixes

* fix management api identifier ([3b3904c](https://www.github.com/instill-ai/api-gateway/commit/3b3904c29f41da4d58942a018c556b64b84fd2ab))
* re-generate krakend config yaml ([fe3de1d](https://www.github.com/instill-ai/api-gateway/commit/fe3de1d9f66ad30349b279cf14d63b762f779c91))
* use jwt from auth backend for inference api ([7878d87](https://www.github.com/instill-ai/api-gateway/commit/7878d878d2a69c3647669fbfdc77cd22d8bb396e))

### [1.5.3](https://www.github.com/instill-ai/api-gateway/compare/v1.5.2...v1.5.3) (2021-03-25)


### Miscellaneous

* move influxdb to be with api-gateway ([a0553df](https://www.github.com/instill-ai/api-gateway/commit/a0553df80c649b982815ae30d80af13e5840e2e8))

### [1.5.2](https://www.github.com/instill-ai/api-gateway/compare/v1.5.1...v1.5.2) (2021-03-25)


### Bug Fixes

* rename KRAKEN_VERSION env to avoid confict with the native krakend env logic ([d1c07bb](https://www.github.com/instill-ai/api-gateway/commit/d1c07bbafec22d1a40aa8d09b1db6ed40acd1ed2))


### Miscellaneous

* bring local krakend.json back ([347b98e](https://www.github.com/instill-ai/api-gateway/commit/347b98e1ac9e0f4c6313a92e5aa75ab14a1261a2))

### [1.5.1](https://www.github.com/instill-ai/api-gateway/compare/v1.5.0...v1.5.1) (2021-03-24)


### Bug Fixes

* remove use of srv dns ([9bf4022](https://www.github.com/instill-ai/api-gateway/commit/9bf4022220d663731832bbcdec2d11d3d7b37ee0))


### Miscellaneous

* introduce host_port env variables for backends ([85124f1](https://www.github.com/instill-ai/api-gateway/commit/85124f1173914e1053837b33bee54b9c58d7893e))

## [1.5.0](https://www.github.com/instill-ai/api-gateway/compare/v1.4.1...v1.5.0) (2021-03-23)


### Features

* add elasticsearch secret and cert tools ([f9f525a](https://www.github.com/instill-ai/api-gateway/commit/f9f525ac7678294aaa9d38e8b69996b0e42cd8ed))
* add filebeat for api-gateway ([2649db1](https://www.github.com/instill-ai/api-gateway/commit/2649db131b880801c985b94eace0f4a8db03eab1))
* add logging stack ([66e9108](https://www.github.com/instill-ai/api-gateway/commit/66e910841c5ee25f4baba921c1ffbc4ec64a4bd3))
* add proxy plugin for logging inference endpoint ([da77cd8](https://www.github.com/instill-ai/api-gateway/commit/da77cd85db1939338985c1a7631e8a23f9ff15ca))
* make krakend flexible config work with env variables ([c309464](https://www.github.com/instill-ai/api-gateway/commit/c30946417d06517c33fb377cbc10e06401aabc66))
* make monitoring stack adopt new compose setup ([b518d1c](https://www.github.com/instill-ai/api-gateway/commit/b518d1c1723bedc563eaeab041c7a0c361c23378))
* refactor docker image build ([5970f07](https://www.github.com/instill-ai/api-gateway/commit/5970f07092996415a1b12b10c0eec1c62a3851cb))


### Miscellaneous

* build in krakend config in the container ([4b57d10](https://www.github.com/instill-ai/api-gateway/commit/4b57d10a01e0451c3fdfd3c1432b5d01223f3084))
* bump up inference-backend to 1.0.3 (golang backend) ([83cafe9](https://www.github.com/instill-ai/api-gateway/commit/83cafe96550f8a155be8ab4ad9130c46cb0b39e7))
* rename krakend folder to api-gateway ([756e049](https://www.github.com/instill-ai/api-gateway/commit/756e04948799bcdf2102699da8470860570c577e))
* rename local dev build image ([4b77aa6](https://www.github.com/instill-ai/api-gateway/commit/4b77aa6715e804850fe2f44d0782cee2244eeb4c))
* replace hardcoded ports with env variables ([1fe6d09](https://www.github.com/instill-ai/api-gateway/commit/1fe6d09ed4a2f4239eaebcdf8e22b79056a99ee4))

### [1.4.1](https://www.github.com/instill-ai/api-gateway/compare/v1.4.0...v1.4.1) (2021-02-24)


### Miscellaneous

* change proxy log format ([025071b](https://www.github.com/instill-ai/api-gateway/commit/025071b1ad240d298a3e21ad0147dc315ded8446))
* trigger ci/cd on all repo change ([0ca963f](https://www.github.com/instill-ai/api-gateway/commit/0ca963f9b76f1758079625133ecf30c8ff9cabb3))

## [1.4.0](https://www.github.com/instill-ai/api-gateway/compare/v1.3.0...v1.4.0) (2021-02-24)


### Features

* add components for monitoring ([cdd99f0](https://www.github.com/instill-ai/api-gateway/commit/cdd99f00fac527e71d59c44985742e4e8b518d51))
* add proxy plugin for logging ([7bc1673](https://www.github.com/instill-ai/api-gateway/commit/7bc16738296df2a88802ac5bd7ac932cfc3722ba))
* adopt KrakenD flexible configuration ([478eb03](https://www.github.com/instill-ai/api-gateway/commit/478eb03ccc59dea406399ef51f70b486227b65a7))


### Bug Fixes

* de-confuse shared volume for ssl cert using depends_on ([a37a149](https://www.github.com/instill-ai/api-gateway/commit/a37a1495851403b3d5a01839c60ca2c446caf441))

## [1.3.0](https://www.github.com/instill-ai/api-gateway/compare/v1.2.0...v1.3.0) (2021-02-06)


### Features

* adopt https completely ([06a10c2](https://www.github.com/instill-ai/api-gateway/commit/06a10c23bc312f34a7f3ede62014383a66a53f5a))
* genereate self-signed tls certificate for inference-backend ([556dc56](https://www.github.com/instill-ai/api-gateway/commit/556dc56f93409130afe6437e15fa1fba7db73f48))


### Bug Fixes

* change port mapping for krakend ([556dac3](https://www.github.com/instill-ai/api-gateway/commit/556dac3b2dda6aee1e11782a3e80032356ce39c4))

## [1.2.0](https://www.github.com/instill-ai/api-gateway/compare/v1.1.0...v1.2.0) (2021-02-03)


### Features

* add debug tools ([58129c3](https://www.github.com/instill-ai/api-gateway/commit/58129c354cfe44d092c31905a66692737f4d0106))
* add self-signed ssl certificate ([8a677c1](https://www.github.com/instill-ai/api-gateway/commit/8a677c1cd4e82f4518865e9ec90e64077f77af47))
* use uwsgi in inference-backend ([0a9f5dc](https://www.github.com/instill-ai/api-gateway/commit/0a9f5dc41538d5faa24659c960887dd4a4567726))


### Miscellaneous

* add staging domain for inference-backend cors ([5b86682](https://www.github.com/instill-ai/api-gateway/commit/5b86682cd2ec6d8092d91c9d76390afd864fc3dd))
* update krakend.json ([711657b](https://www.github.com/instill-ai/api-gateway/commit/711657bbabe144ef8b9fedda825e674f69a99fc6))

## [1.1.0](https://www.github.com/instill-ai/api-gateway/compare/v1.0.0...v1.1.0) (2021-01-29)


### Features

* add detection endpoint ([1cf2fb4](https://www.github.com/instill-ai/api-gateway/commit/1cf2fb455be5e3e9cafa961bae935e4329567e7b))
* add docker-compose.yml ([3080898](https://www.github.com/instill-ai/api-gateway/commit/3080898e0ab5493f776317e074ab8b2c076f2657))

## 1.0.0 (2021-01-29)


### Features

* add dockerfile ([aa05c62](https://www.github.com/instill-ai/api-gateway/commit/aa05c62b4b7b97aeb2102fbf2d5dbb0aa40dece2))
* add krakend.json ([50275b4](https://www.github.com/instill-ai/api-gateway/commit/50275b41ab50c7bc45e4b59bc14639feb378c393))


### CI/CD

* add ci/cd workflows ([d8d12c0](https://www.github.com/instill-ai/api-gateway/commit/d8d12c001b8d7a10d36a0c1c4758af3ef8d0bd6d))
