# Changelog

## [0.30.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.29.0-beta...v0.30.0-beta) (2024-11-18)


### Features

* **vdp:** expose webhook dispatcher endpoint ([#266](https://github.com/instill-ai/api-gateway/issues/266)) ([f00deed](https://github.com/instill-ai/api-gateway/commit/f00deedc9f9213d31c0d404bf315218425ba87be))


### Bug Fixes

* **blob:** blob support authn by jwt token ([#267](https://github.com/instill-ai/api-gateway/issues/267)) ([9fac68c](https://github.com/instill-ai/api-gateway/commit/9fac68cf5b52b06205726dc2ae201298b3dfe835))

## [0.29.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.28.1-beta...v0.29.0-beta) (2024-11-05)


### Features

* **blob:** relay file download and upload to minio ([#261](https://github.com/instill-ai/api-gateway/issues/261)) ([ec27223](https://github.com/instill-ai/api-gateway/commit/ec2722362f8db304032581040969a08405750295))
* **run:** pipeline/model runs and metrics APIs ([#265](https://github.com/instill-ai/api-gateway/issues/265)) ([c5f1a8d](https://github.com/instill-ai/api-gateway/commit/c5f1a8d765384493865ea7eacef5c36a227ac65b))


### Bug Fixes

* **blob:** build the blob plugin in image ([#263](https://github.com/instill-ai/api-gateway/issues/263)) ([24fc718](https://github.com/instill-ai/api-gateway/commit/24fc718aabcffe63b9194fff256ae2db1a370b08))

## [0.28.1-beta](https://github.com/instill-ai/api-gateway/compare/v0.28.0-beta...v0.28.1-beta) (2024-10-23)


### Bug Fixes

* **metric:** reintroduce /metrics/vdp/pipeline/triggers endpoint ([#259](https://github.com/instill-ai/api-gateway/issues/259)) ([6624dcd](https://github.com/instill-ai/api-gateway/commit/6624dcd48c2337081d4c5efed9c8d2031e7878d9))

## [0.28.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.27.0-beta...v0.28.0-beta) (2024-10-22)


### Features

* **blob:** add blob endpoint ([#255](https://github.com/instill-ai/api-gateway/issues/255)) ([7734452](https://github.com/instill-ai/api-gateway/commit/7734452b36743a0b5f6485442a5ca4e0f8a44e7e))


### Bug Fixes

* **connection:** fix GetNamespaceConnection route ([#257](https://github.com/instill-ai/api-gateway/issues/257)) ([7f18715](https://github.com/instill-ai/api-gateway/commit/7f18715d4a378a0fede773633c5542b237fd64db))

## [0.27.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.26.0-beta...v0.27.0-beta) (2024-10-04)


### Features

* **artifact:** add chunk uid in query string ([#252](https://github.com/instill-ai/api-gateway/issues/252)) ([0180e75](https://github.com/instill-ai/api-gateway/commit/0180e758c5d16c45ef411ead9646dd25720a9e32))

## [0.26.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.25.0-beta...v0.26.0-beta) (2024-09-24)


### Miscellaneous Chores

* release v0.26.0-beta ([5a85791](https://github.com/instill-ai/api-gateway/commit/5a8579159ade275da94f1b9b1e37814ffcab0cc1))

## [0.25.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.24.0-beta...v0.25.0-beta) (2024-09-10)


### Features

* **pipeline:** add connection create and get endpoints ([#246](https://github.com/instill-ai/api-gateway/issues/246)) ([28e7cdc](https://github.com/instill-ai/api-gateway/commit/28e7cdc8f16ff13f758b9f62a05f72e0c6259bcf))
* **pipeline:** add integration read endpoints ([#244](https://github.com/instill-ai/api-gateway/issues/244)) ([2e4d8fa](https://github.com/instill-ai/api-gateway/commit/2e4d8fa401e2d394dd371475a001c05e12195506))
* **pipeline:** expose integration endpoints ([#248](https://github.com/instill-ai/api-gateway/issues/248)) ([b02ec58](https://github.com/instill-ai/api-gateway/commit/b02ec5896a93c3b338e6db15a90dd55df83359f8))
* **vdp:** add grpc pipeline run logging endpoints ([#247](https://github.com/instill-ai/api-gateway/issues/247)) ([795aeb0](https://github.com/instill-ai/api-gateway/commit/795aeb003f68772b0849c2dd983680b32a419b92))

## [0.24.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.23.0-beta...v0.24.0-beta) (2024-08-26)


### Features

* **catalog:** add conversations and messages endpoints ([#234](https://github.com/instill-ai/api-gateway/issues/234)) ([7734662](https://github.com/instill-ai/api-gateway/commit/7734662665fd2b9d2519ce9606c2f51ca934178d))
* **model:** added model run logging endpoint ([#231](https://github.com/instill-ai/api-gateway/issues/231)) ([e7c3232](https://github.com/instill-ai/api-gateway/commit/e7c323247e23ab782062d318fa4a62ca521f687f))
* **model:** fixed model run logging endpoint prefix ([#235](https://github.com/instill-ai/api-gateway/issues/235)) ([1051b64](https://github.com/instill-ai/api-gateway/commit/1051b6489054d89fabf14bfee5e277fc0d639750))
* **model:** fixed model run logging endpoint query param ([#233](https://github.com/instill-ai/api-gateway/issues/233)) ([2d11f99](https://github.com/instill-ai/api-gateway/commit/2d11f997a708ba95b520a6060bd068082058fba6))
* **vdp:** added pipeline run logging endpoints ([#237](https://github.com/instill-ai/api-gateway/issues/237)) ([c0a52d9](https://github.com/instill-ai/api-gateway/commit/c0a52d9e494d276bf1e2bb70bcdc7a512587c9e6))
* **vdp:** support authentication for streaming API ([#238](https://github.com/instill-ai/api-gateway/issues/238)) ([a99d723](https://github.com/instill-ai/api-gateway/commit/a99d7230caa12cc2c1ee8867dad867190cde85cc))
* **vdp:** use the `Accept` header to identify and handle `text/event-stream` ([#242](https://github.com/instill-ai/api-gateway/issues/242)) ([3891071](https://github.com/instill-ai/api-gateway/commit/389107180eb5930dea2c8db964b1c4f69260c70c))


### Bug Fixes

* **model:** fix wrong version and add missing grpc endpoint ([#236](https://github.com/instill-ai/api-gateway/issues/236)) ([78cb564](https://github.com/instill-ai/api-gateway/commit/78cb5645b2144014d3295d6f2b4d30998afea89c))

## [0.23.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.22.0-beta...v0.23.0-beta) (2024-08-13)


### Features

* **catalog:** add file catalog and question answering endpoint ([#226](https://github.com/instill-ai/api-gateway/issues/226)) ([a5c85dc](https://github.com/instill-ai/api-gateway/commit/a5c85dc24221613404f25472949c6187fdbd4f6a))
* **env:** update KrakenD to v2.7.0 ([#230](https://github.com/instill-ai/api-gateway/issues/230)) ([4c6144b](https://github.com/instill-ai/api-gateway/commit/4c6144b1c76388919d3568958b64bd13faaa8cbc))


### Bug Fixes

* **catalog:** fix typo in file catalog api ([#229](https://github.com/instill-ai/api-gateway/issues/229)) ([05fea14](https://github.com/instill-ai/api-gateway/commit/05fea144447d82aa2b482fce13655f4e0e237b3d))
* **registry:** fix wrong namespace format comparison ([#228](https://github.com/instill-ai/api-gateway/issues/228)) ([dbe1818](https://github.com/instill-ai/api-gateway/commit/dbe18186b1478fe7593ab50b27bacb7c4f09d4e2))

## [0.22.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.21.0-beta...v0.22.0-beta) (2024-08-01)


### Features

* **catalog:** rename endpoint from kb to catalog ([#224](https://github.com/instill-ai/api-gateway/issues/224)) ([71e7775](https://github.com/instill-ai/api-gateway/commit/71e77755c795c72561ee7c74df6a2718e00e7d2d))

## [0.21.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.20.0-beta...v0.21.0-beta) (2024-07-30)


### Features

* **artifact:** use namespace and add similary search api ([#214](https://github.com/instill-ai/api-gateway/issues/214)) ([97074e8](https://github.com/instill-ai/api-gateway/commit/97074e8e7ac263689f2312f96db3be30e8318d55))
* **mgmt:** use unified dashboard endpoints ([#218](https://github.com/instill-ai/api-gateway/issues/218)) ([f3522ac](https://github.com/instill-ai/api-gateway/commit/f3522ac61eacbac50fcb31e9dea373785eade57e))
* **registry:** adopt namespace endpoint ([#219](https://github.com/instill-ai/api-gateway/issues/219)) ([ac2ce07](https://github.com/instill-ai/api-gateway/commit/ac2ce07f750558145b50552a057eb03f845e5093))
* **vdp,model:** expose namespace endpoints ([#217](https://github.com/instill-ai/api-gateway/issues/217)) ([386dc58](https://github.com/instill-ai/api-gateway/commit/386dc58eaa7cf312307e989cf9cafd18e1e98b3f))
* **vdp:** expose run-on-event endpoints ([#221](https://github.com/instill-ai/api-gateway/issues/221)) ([55d8cab](https://github.com/instill-ai/api-gateway/commit/55d8cab5462f7aa3ac1a53c7b31627ccca879b58))


### Bug Fixes

* **artifact:** fix gRPC endpoints not correctly exposed ([#223](https://github.com/instill-ai/api-gateway/issues/223)) ([a32a1e0](https://github.com/instill-ai/api-gateway/commit/a32a1e0c7800b8252019834d910a36cbe5b7c54c))
* **kb:** upload file timeout ([#222](https://github.com/instill-ai/api-gateway/issues/222)) ([48ba2f5](https://github.com/instill-ai/api-gateway/commit/48ba2f560148f46ac011cdc6c797bdbe3647cfdd))
* **mgmt:** restore pipeline dashboard enpoints ([#220](https://github.com/instill-ai/api-gateway/issues/220)) ([82e79bb](https://github.com/instill-ai/api-gateway/commit/82e79bb56d9cd2ab828b9804db1b105ae3918efe))

## [0.20.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.19.0-beta...v0.20.0-beta) (2024-07-15)


### Features

* **cors:** allow Instill-Use-SSE header ([#209](https://github.com/instill-ai/api-gateway/issues/209)) ([296740f](https://github.com/instill-ai/api-gateway/commit/296740f3c5d5bb8cce6903364d24aac64e788c62))
* **vdp:** add server plugin to enable SSE streaming and enable endpoints and configuration ([#202](https://github.com/instill-ai/api-gateway/issues/202)) ([7d41024](https://github.com/instill-ai/api-gateway/commit/7d4102485b3a96e6b4ddb9c563133a375ca31c76))
* **vdp:** expose pipeline profile image endpoint ([#211](https://github.com/instill-ai/api-gateway/issues/211)) ([006f016](https://github.com/instill-ai/api-gateway/commit/006f016148db732892485737c2705473591d9b46))

## [0.19.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.18.0-beta...v0.19.0-beta) (2024-07-01)


### Features

* **model:** add model version deletion endpoints ([#203](https://github.com/instill-ai/api-gateway/issues/203)) ([77714f4](https://github.com/instill-ai/api-gateway/commit/77714f42990e3d762fffb473aa2ecb7506f9e764))
* **pipeline:** add requester header to CORS ([#207](https://github.com/instill-ai/api-gateway/issues/207)) ([9cbe156](https://github.com/instill-ai/api-gateway/commit/9cbe156e16d5deff6064b8a52cb72fbbf5fa3b00))
* **pipeline:** add requester UID header ([#205](https://github.com/instill-ai/api-gateway/issues/205)) ([f56379e](https://github.com/instill-ai/api-gateway/commit/f56379ef126387e3d824d68db1f156551b4f2c74))
* **vdp:** expose endpoints for cloning pipeline release ([#206](https://github.com/instill-ai/api-gateway/issues/206)) ([d813ad6](https://github.com/instill-ai/api-gateway/commit/d813ad6a2d0c2c2bb3c35e15036a1f91f5ddc097))

## [0.18.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.17.0-beta...v0.18.0-beta) (2024-06-17)


### Features

* **endpoints:** use camelCase for query strings ([#199](https://github.com/instill-ai/api-gateway/issues/199)) ([bb3a943](https://github.com/instill-ai/api-gateway/commit/bb3a943cfd0ca1f8eb9e88245239c84fee76ef4b))
* **kb:** support knowledge base file related api ([#194](https://github.com/instill-ai/api-gateway/issues/194)) ([832c55f](https://github.com/instill-ai/api-gateway/commit/832c55f3e9c07a414997b4ea26e7f22ee329e436))
* **plugin:** adopt multi-auth plugin to adopt the changes in /login endpoint ([#197](https://github.com/instill-ai/api-gateway/issues/197)) ([b4d1897](https://github.com/instill-ai/api-gateway/commit/b4d1897b01b185244d85b57004a0be33bff0b382))

## [0.17.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.16.0-beta...v0.17.0-beta) (2024-06-06)


### Features

* **model:** add order_by field for list model endpoints ([#186](https://github.com/instill-ai/api-gateway/issues/186)) ([7486703](https://github.com/instill-ai/api-gateway/commit/748670372e6180028f6a68d519ed9c9589061719))
* **model:** support trigger latest model version ([#185](https://github.com/instill-ai/api-gateway/issues/185)) ([bbf250d](https://github.com/instill-ai/api-gateway/commit/bbf250d009c1422c5c034afa49747c01bcfe54f9))


### Bug Fixes

* **kb:** endpoint conflict ([#189](https://github.com/instill-ai/api-gateway/issues/189)) ([5051f09](https://github.com/instill-ai/api-gateway/commit/5051f09e2a25835e18ef52e6a3265cc31b8eeb17))
* **registry:** fix org membership check in registry plugin ([#182](https://github.com/instill-ai/api-gateway/issues/182)) ([5370041](https://github.com/instill-ai/api-gateway/commit/53700418e56451641c60838b97276899580412d4))

## [0.16.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.15.0-beta...v0.16.0-beta) (2024-05-15)


### Features

* **model:** support listing available regions for model deployment ([#178](https://github.com/instill-ai/api-gateway/issues/178)) ([37333d1](https://github.com/instill-ai/api-gateway/commit/37333d1e644d405887de1f83b5d76ed388c5185b))
* **pipeline:** add order_by param for pipeline endpoints ([#181](https://github.com/instill-ai/api-gateway/issues/181)) ([c66b67d](https://github.com/instill-ai/api-gateway/commit/c66b67dfdad0b8a1a1efc8a8face476c43f3df90))

## [0.15.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.14.0-beta...v0.15.0-beta) (2024-05-07)


### Features

* **core:** remove basePath for core, vdp, model and artifact endpoints ([#175](https://github.com/instill-ai/api-gateway/issues/175)) ([8766224](https://github.com/instill-ai/api-gateway/commit/876622497f4779a185d5b78e3564762669d2a447))

## [0.14.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.13.0-beta...v0.14.0-beta) (2024-04-24)


### Features

* **mgmt:** add credit public endpoints ([#168](https://github.com/instill-ai/api-gateway/issues/168)) ([8643c26](https://github.com/instill-ai/api-gateway/commit/8643c26a220524bf99e6f8eddef4c23b7e12dcb4))
* **vdp:** adjust user secrets endpoint ([#172](https://github.com/instill-ai/api-gateway/issues/172)) ([f76992b](https://github.com/instill-ai/api-gateway/commit/f76992b00453ec1dd8f952d89228808ff8158d24))
* **vdp:** expose pipeline secrets endpoints ([#164](https://github.com/instill-ai/api-gateway/issues/164)) ([0ade2f8](https://github.com/instill-ai/api-gateway/commit/0ade2f88789a73da15094bc1939e8464db54bb3a))
* **vdp:** move connector/operator definition endpoints to auth section ([#174](https://github.com/instill-ai/api-gateway/issues/174)) ([0f5e0d1](https://github.com/instill-ai/api-gateway/commit/0f5e0d1f8958e713f3b46848e9356ddc21bbc5d9))

## [0.13.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.12.0-beta...v0.13.0-beta) (2024-04-11)


### Features

* **artifact:** create tag on successful push ([#153](https://github.com/instill-ai/api-gateway/issues/153)) ([2bfe85a](https://github.com/instill-ai/api-gateway/commit/2bfe85ad8b1dc71ef2cad8d41093fba47709e285))
* **artifact:** deploy model on push completion ([#152](https://github.com/instill-ai/api-gateway/issues/152)) ([19abc68](https://github.com/instill-ai/api-gateway/commit/19abc680968caa3ff9a066f33cbe198e2a5b9170))
* **artifact:** expose public Artifact endpoints ([#149](https://github.com/instill-ai/api-gateway/issues/149)) ([02ac5c5](https://github.com/instill-ai/api-gateway/commit/02ac5c5d4fa71bb40bd85120a477720cc774923d))
* **mgmt:** add endpoints for user and organization avatars ([#162](https://github.com/instill-ai/api-gateway/issues/162)) ([bbcd8c8](https://github.com/instill-ai/api-gateway/commit/bbcd8c8115022f83d0ef157fefab9f258ed2d827))
* **model:** add model async trigger ([#159](https://github.com/instill-ai/api-gateway/issues/159)) ([35f3165](https://github.com/instill-ai/api-gateway/commit/35f3165f36798a307d2bcc1f3cb8bb632db05f8d))
* **model:** add namespace check and adopt latest protobuf ([#156](https://github.com/instill-ai/api-gateway/issues/156)) ([64fcd6e](https://github.com/instill-ai/api-gateway/commit/64fcd6e31cda47ed6c7f8c60fbdcdb9dd76a37ed))


### Bug Fixes

* **registry:** capture NotFound response in namespace check ([#160](https://github.com/instill-ai/api-gateway/issues/160)) ([7ed29dd](https://github.com/instill-ai/api-gateway/commit/7ed29ddcc1cae946cb0eab3b1f0e3193d10057ee))
* **registry:** handle errors complying with the V2 API ([#158](https://github.com/instill-ai/api-gateway/issues/158)) ([3fb5532](https://github.com/instill-ai/api-gateway/commit/3fb5532c7de170d0ace28a78d8ae78e758b7daed))

## [0.12.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.11.0-beta...v0.12.0-beta) (2024-03-30)


### Features

* **registry:** add registry proxy plugin ([#143](https://github.com/instill-ai/api-gateway/issues/143)) ([97c280c](https://github.com/instill-ai/api-gateway/commit/97c280cbdcb615c71ab29a90c502f7f97d26426e))

## [0.11.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.10.0-beta...v0.11.0-beta) (2024-02-29)


### Features

* **vdp:** add component definition list endpoint ([#137](https://github.com/instill-ai/api-gateway/issues/137)) ([8e459b7](https://github.com/instill-ai/api-gateway/commit/8e459b7c3ef862d17c219afbc6064f8ff0c2d45c))


### Bug Fixes

* **vdp:** implement offset pagination in component list endpoint ([#139](https://github.com/instill-ai/api-gateway/issues/139)) ([a6b7566](https://github.com/instill-ai/api-gateway/commit/a6b75667a9909341009389421a2adbe0881bf087))

## [0.10.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.9.0-beta...v0.10.0-beta) (2024-02-05)


### Features

* **mgmt:** refactor API for `GET` and `PATCH` authenticated user ([#133](https://github.com/instill-ai/api-gateway/issues/133)) ([d4c3268](https://github.com/instill-ai/api-gateway/commit/d4c3268219e1db7b1956b420517ed14c4b9cf69c))
* **model:** add model organization endpoints ([#135](https://github.com/instill-ai/api-gateway/issues/135)) ([b653945](https://github.com/instill-ai/api-gateway/commit/b6539454d06bf3fd97300cc9829b91cb41de1fed))

## [0.9.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.8.0-beta...v0.9.0-beta) (2024-01-29)


### Features

* **vdp:** add `CheckName` endpoint ([#132](https://github.com/instill-ai/api-gateway/issues/132)) ([e5ad528](https://github.com/instill-ai/api-gateway/commit/e5ad528be29565044566b2dda1ebbad52fdd112a))
* **vdp:** add `CloneUserPipeline` and `CloneOrganizationPipeline` endpoints ([#131](https://github.com/instill-ai/api-gateway/issues/131)) ([633967f](https://github.com/instill-ai/api-gateway/commit/633967fe67f81a89aeb64d05e7b4d641702607ca))
* **vdp:** add visibility query param for list pipelines endpoints ([#130](https://github.com/instill-ai/api-gateway/issues/130)) ([d664bc3](https://github.com/instill-ai/api-gateway/commit/d664bc3d4d44d329e408a2afc7a4fdfc9d41953f))

## [0.8.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.7.1-beta...v0.8.0-beta) (2024-01-17)


### Features

* **vdp:** propagate `visibility` query param for GET /pipelines endpoint ([#124](https://github.com/instill-ai/api-gateway/issues/124)) ([6c3bb64](https://github.com/instill-ai/api-gateway/commit/6c3bb64748d7b43a1d321dc8971b9d41ed960012))

## [0.7.1-beta](https://github.com/instill-ai/api-gateway/compare/v0.7.0-beta...v0.7.1-beta) (2023-12-29)


### Bug Fixes

* **dockerfile:** build Krakend from source to fix golang vesion mismatch ([#119](https://github.com/instill-ai/api-gateway/issues/119)) ([6c9edc2](https://github.com/instill-ai/api-gateway/commit/6c9edc2eb3a58e812d7d6d86b28776ebecdb742d))
* **dockerfile:** fix cross compiling issues ([#122](https://github.com/instill-ai/api-gateway/issues/122)) ([558b99e](https://github.com/instill-ai/api-gateway/commit/558b99e186fbd309f83f1252d6cf11a89f6e4b58))

## [0.7.0-beta](https://github.com/instill-ai/api-gateway/compare/v0.6.0-alpha...v0.7.0-beta) (2023-12-15)


### Features

* **core,vdp:** upgrade endpoints version from v1alpha to v1beta ([#116](https://github.com/instill-ai/api-gateway/issues/116)) ([aaa153f](https://github.com/instill-ai/api-gateway/commit/aaa153ff33adc727233aed04938370652c2a1a36))
* **plugin:** inject visitor id for public endpoints ([#114](https://github.com/instill-ai/api-gateway/issues/114)) ([ea6ebfa](https://github.com/instill-ai/api-gateway/commit/ea6ebfa2af83b6b54fc90ac0685777de68e0f556))
* **vdp:** add organization pipeline and connector endpoints ([#112](https://github.com/instill-ai/api-gateway/issues/112)) ([aa08892](https://github.com/instill-ai/api-gateway/commit/aa08892575e987fd5555a3908c0d39c00497f5c7))


### Miscellaneous Chores

* **release:** release v0.7.0-beta ([a91aa20](https://github.com/instill-ai/api-gateway/commit/a91aa207002c677955c7c247081196e134a53dc4))

## [0.6.0-alpha](https://github.com/instill-ai/api-gateway/compare/v0.5.3-alpha...v0.6.0-alpha) (2023-11-22)


### Features

* **core:** add organization endpoints ([#109](https://github.com/instill-ai/api-gateway/issues/109)) ([2a7f2a6](https://github.com/instill-ai/api-gateway/commit/2a7f2a604ba1bdc3c70df72ce5ab39f972a7baa1))

## [0.5.3-alpha](https://github.com/instill-ai/api-gateway/compare/v0.5.2-alpha...v0.5.3-alpha) (2023-11-11)


### Miscellaneous Chores

* **release:** release v0.5.3-alpha ([b2b9d25](https://github.com/instill-ai/api-gateway/commit/b2b9d25f4c34699c5ca4e66e8034361dc58e1e30))

## [0.5.2-alpha](https://github.com/instill-ai/api-gateway/compare/v0.5.1-alpha...v0.5.2-alpha) (2023-10-27)


### Miscellaneous Chores

* **release:** release v0.5.2-alpha ([2f3cb5c](https://github.com/instill-ai/api-gateway/commit/2f3cb5ce111029714f120dc112bc5b0c2cdb818d))

## [0.5.1-alpha](https://github.com/instill-ai/api-gateway/compare/v0.5.0-alpha...v0.5.1-alpha) (2023-10-13)


### Bug Fixes

* **plugin:** implement KrakenD grpc proxy server plugin to fix HTTP/2 Trailer issue ([#94](https://github.com/instill-ai/api-gateway/issues/94)) ([8c82733](https://github.com/instill-ai/api-gateway/commit/8c8273314719136ffd341ecac8fa7a613ed49eec))

## [0.5.0-alpha](https://github.com/instill-ai/api-gateway/compare/v0.4.0-alpha...v0.5.0-alpha) (2023-09-26)


### Features

* **auth:** add auth/signer and auth/validator ([#85](https://github.com/instill-ai/api-gateway/issues/85)) ([5e13fc2](https://github.com/instill-ai/api-gateway/commit/5e13fc20b8224e3c9afaf23e7bf09379a4fe22ba))
* **plugin:** add multi_auth plugin to support `api_token` authentication and basic_auth ([#87](https://github.com/instill-ai/api-gateway/issues/87)) ([a482904](https://github.com/instill-ai/api-gateway/commit/a4829046f3d296f3e00cec1afea4a084ee5f52b5))


### Bug Fixes

* **auth:** fix api gateway `jwx` tool failed in arm64 ([#89](https://github.com/instill-ai/api-gateway/issues/89)) ([b76f85f](https://github.com/instill-ai/api-gateway/commit/b76f85f1d118f983b28057b8b9869f6d194e4043))
* **base:** fix pipeline release `setDefault` endpoint ([#93](https://github.com/instill-ai/api-gateway/issues/93)) ([d1b33cd](https://github.com/instill-ai/api-gateway/commit/d1b33cd599bfda18b6b62f4e5f9fbe8c2197f647))

## [0.4.0-alpha](https://github.com/instill-ai/api-gateway/compare/v0.3.2-alpha...v0.4.0-alpha) (2023-09-13)


### Miscellaneous Chores

* **release:** release v0.4.0-alpha ([74585ef](https://github.com/instill-ai/api-gateway/commit/74585ef620b33f263464444e68e7110da31f5e21))

## [0.3.2-alpha](https://github.com/instill-ai/api-gateway/compare/v0.3.1-alpha...v0.3.2-alpha) (2023-08-03)


### Miscellaneous Chores

* **release:** release v0.3.2-alpha ([9fbcd3a](https://github.com/instill-ai/api-gateway/commit/9fbcd3a3498e4e8676112eb294028b68c68a5b71))

## [0.3.1-alpha](https://github.com/instill-ai/api-gateway/compare/v0.3.0-alpha...v0.3.1-alpha) (2023-07-20)


### Miscellaneous Chores

* **release:** release v0.3.1-alpha ([8cc0fa8](https://github.com/instill-ai/api-gateway/commit/8cc0fa8fd6ecb24943ca057c02ec0c4461cb166f))

## [0.3.0-alpha](https://github.com/instill-ai/api-gateway/compare/v0.2.8-alpha...v0.3.0-alpha) (2023-07-09)


### Miscellaneous Chores

* **release:** release v0.3.0-alpha ([bd0d73e](https://github.com/instill-ai/api-gateway/commit/bd0d73ed884728b2ff3c4b78f91dd39aac005cac))

## [0.2.8-alpha](https://github.com/instill-ai/api-gateway/compare/v0.2.7-alpha...v0.2.8-alpha) (2023-06-20)


### Miscellaneous Chores

* **release:** release 0.2.8-alpha ([486d600](https://github.com/instill-ai/api-gateway/commit/486d600bcc1ed9556965d4aef560f651e6190186))

## [0.2.7-alpha](https://github.com/instill-ai/api-gateway/compare/v0.2.6-alpha...v0.2.7-alpha) (2023-06-02)


### Miscellaneous Chores

* **release:** release v0.2.7-alpha ([3624550](https://github.com/instill-ai/api-gateway/commit/3624550522422967fa76d3a465f444ca9aaa7b8f))

## [0.2.6-alpha](https://github.com/instill-ai/api-gateway/compare/v0.2.5-alpha...v0.2.6-alpha) (2023-05-11)


### Bug Fixes

* refactor model backend endpoint ([#36](https://github.com/instill-ai/api-gateway/issues/36)) ([6a3283f](https://github.com/instill-ai/api-gateway/commit/6a3283fc86a83e2897e405b9b611087326cb9206))


### Miscellaneous Chores

* **release:** release v0.2.6-alpha ([929dee8](https://github.com/instill-ai/api-gateway/commit/929dee8b0ca0eff505f4e31587b5f13b54be368b))

## [0.2.5-alpha](https://github.com/instill-ai/api-gateway/compare/v0.2.4-alpha...v0.2.5-alpha) (2023-04-14)


### Bug Fixes

* **endpoints:** update service name ([#33](https://github.com/instill-ai/api-gateway/issues/33)) ([cdf8e4f](https://github.com/instill-ai/api-gateway/commit/cdf8e4f1f37a622b7581d4751e9cc4fd4eb438b8))

## [0.2.4-alpha](https://github.com/instill-ai/api-gateway/compare/v0.2.3-alpha...v0.2.4-alpha) (2023-04-09)


### Miscellaneous Chores

* release v0.2.4-alpha ([797410a](https://github.com/instill-ai/api-gateway/commit/797410a149409c1e83086975425f61f76adf4cb4))

## [0.2.3-alpha](https://github.com/instill-ai/api-gateway/compare/v0.2.2-alpha...v0.2.3-alpha) (2023-03-26)


### Miscellaneous Chores

* release v0.2.3-alpha ([bbfa661](https://github.com/instill-ai/api-gateway/commit/bbfa661e85041eabb07efe2c3cbda103c2673b89))

## [0.2.2-alpha](https://github.com/instill-ai/api-gateway/compare/v0.2.1-alpha...v0.2.2-alpha) (2023-02-23)


### Miscellaneous Chores

* release v0.2.2-alpha ([d7f3dd0](https://github.com/instill-ai/api-gateway/commit/d7f3dd0c2f6dc4b7b009923beaa21973e73936dd))

## [0.2.1-alpha](https://github.com/instill-ai/api-gateway/compare/v0.2.0-alpha...v0.2.1-alpha) (2023-02-10)


### Bug Fixes

* fix endpoints and remove jwt token configuration ([#13](https://github.com/instill-ai/api-gateway/issues/13)) ([7b35a60](https://github.com/instill-ai/api-gateway/commit/7b35a6075475e3f5b3d5361b3e7e05bb6c3923e0))
* update configuration for headers allowed to reach the backend ([#16](https://github.com/instill-ai/api-gateway/issues/16)) ([a4f66b2](https://github.com/instill-ai/api-gateway/commit/a4f66b200105b7f9b9aa37e81efeeaafc818d52a))

## [0.2.0-alpha](https://github.com/instill-ai/api-gateway/compare/v0.1.1-alpha...v0.2.0-alpha) (2023-01-15)


### Features

* enable h2c ([89db3e9](https://github.com/instill-ai/api-gateway/commit/89db3e90f5f1c2cb2ff4bc46e68e6825a24ccb8d))

## [0.1.1-alpha](https://github.com/instill-ai/api-gateway/compare/v0.1.0-alpha...v0.1.1-alpha) (2023-01-03)


### Miscellaneous Chores

* release v0.1.1-alpha ([6f12078](https://github.com/instill-ai/api-gateway/commit/6f12078269e1379ca3991a27eecd664fb6c1fa82))

## [0.1.0-alpha](https://github.com/instill-ai/api-gateway/compare/v0.0.0-alpha...v0.1.0-alpha) (2022-12-30)


### Features

* add pipeline and connector backend ([82cc726](https://github.com/instill-ai/api-gateway/commit/82cc726948daa4df439911134d9e5dd942b28df8))
* **model:** add operation endpoints ([014bd4a](https://github.com/instill-ai/api-gateway/commit/014bd4a3b1b023d8cd8e707c1546fb5954a8a7b0))
