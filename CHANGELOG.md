# Changelog

## [0.3.0](https://github.com/Otard95/ed-expedition/compare/v0.2.1...v0.3.0) (2025-12-25)


### Features

* Add transaction system for atomic multi-file writes ([e0d1781](https://github.com/Otard95/ed-expedition/commit/e0d1781de73b68a73aee70d92c5566e22e061be5))
* **expedition:** Support unknown start location with -1 index ([b985673](https://github.com/Otard95/ed-expedition/commit/b9856739b6a76cc6689a12dc326fe8a38c4a4b94))
* **expedition:** Track current jump in history ([2105cbf](https://github.com/Otard95/ed-expedition/commit/2105cbf737f3d64b25e19690d0d248e8cae8199e))
* **fuel:** Add FuelAlertHandler for toast notifications ([f004017](https://github.com/Otard95/ed-expedition/commit/f00401702af5f0c10811930709408a3fb1608fe4))
* **fuel:** Add real-time fuel tracking and alerts ([01ae571](https://github.com/Otard95/ed-expedition/commit/01ae571705c1097236f470456343f1f0dc94307b))
* **fuel:** Add tiered fuel warnings with scoopable-aware alerts ([c067745](https://github.com/Otard95/ed-expedition/commit/c067745ae18d5ed23579705e09854ba117714ace))
* **journal:** Add watching and events for Status.json ([c388e8a](https://github.com/Otard95/ed-expedition/commit/c388e8aedb5f0b2bb19161d64343c48e0a3e5ee4))
* **jump-repl:** Add fuel level tracking and scooping simulation ([daa646d](https://github.com/Otard95/ed-expedition/commit/daa646ddd4bad7f18a0d73af24bbcc32cd31fbb1))
* **repl:** Add fuel and scooping commands for testing ([350b67e](https://github.com/Otard95/ed-expedition/commit/350b67e967fc0c847ddc8ccd51ae8d10acdeb218))
* **routes:** Enhance fuel display with actual vs expected comparison ([7a41f3c](https://github.com/Otard95/ed-expedition/commit/7a41f3c9de4794aeecbb6a9ef1476b987b9c6a8b))
* **table:** Add tooltip support for column headers ([efa3921](https://github.com/Otard95/ed-expedition/commit/efa3921acdec803c0d38af199575458c7c3ded61))
* **toasts:** Add back to index button in toast test view ([d691135](https://github.com/Otard95/ed-expedition/commit/d691135519550469ec321ab01ebd88eeff4780e6))
* **toasts:** Add toast notification system ([#12](https://github.com/Otard95/ed-expedition/issues/12)) ([59ce42a](https://github.com/Otard95/ed-expedition/commit/59ce42a795df17743f53f636b5f640503563016a))


### Bug Fixes

* **journal:** Improve Status.json error handling and logging ([f0df07a](https://github.com/Otard95/ed-expedition/commit/f0df07a6c42b2a5d0c19c1a894b65ecf8ca81e05))
* **routes:** Update fuel display for all matching systems ([c235663](https://github.com/Otard95/ed-expedition/commit/c235663d5a8e5c1177e81bb4e2e3cd1c2eff68a8))
* **tooltip:** Use fixed positioning to prevent clipping ([3e6b280](https://github.com/Otard95/ed-expedition/commit/3e6b280d9446f1ce5a43d59e2d8f7a33620edb7a))

## [0.2.1](https://github.com/Otard95/ed-expedition/compare/v0.2.0...v0.2.1) (2025-12-21)


### Bug Fixes

* **app-state:** No panic on failed save. Log should be more helpful ([402ac90](https://github.com/Otard95/ed-expedition/commit/402ac90d2a9279ad42e7666a19f5004a4e3491c4))
* **journal:** Correct log name ([2631567](https://github.com/Otard95/ed-expedition/commit/26315673afb426a7e70475c392117b8271b1c7da))
* **journal:** Watcher needs to handle event's within the same second ([a0986c3](https://github.com/Otard95/ed-expedition/commit/a0986c3fc7159a888a520a99fc56f9cf21567537))

## [0.2.0](https://github.com/Otard95/ed-expedition/compare/v0.1.4...v0.2.0) (2025-12-21)


### Features

* **active-view:** sticky stats card, additional metrics, and end expedition ([6fc0150](https://github.com/Otard95/ed-expedition/commit/6fc0150fd5125d059eb7b8aa2ec716f8bdf67243))
* add IntersectionObserver component ([2f813fa](https://github.com/Otard95/ed-expedition/commit/2f813fa0d77bfbde009c217afae437f5cb77163b))
* auto-copy next system name to clipboard after each jump ([d6b23d3](https://github.com/Otard95/ed-expedition/commit/d6b23d384f6201c9826404a8b87c1b42bea199e0))
* **css:** add semantic CSS variables for consistency ([e67787c](https://github.com/Otard95/ed-expedition/commit/e67787cf2f27dffb378931d53c0c45c99b337589))
* **Tooltip:** add direction, size, and nowrap props ([e6e78cf](https://github.com/Otard95/ed-expedition/commit/e6e78cf28b044eeea4e700f2e34570ede34874c7))


### Bug Fixes

* LoadActiveExpedition signature and journal sync condition ([568aeb9](https://github.com/Otard95/ed-expedition/commit/568aeb937b65ba7736fe23e71e5c7f9bb6cc62d1))
* **Modal:** enforce content-controls-size architecture ([b871132](https://github.com/Otard95/ed-expedition/commit/b8711320e9c715a06517a76f8dd9a7689467f2f9))
* **table:** prevent header wrapping and optimize numeric column widths ([4d79804](https://github.com/Otard95/ed-expedition/commit/4d798043203444676103e0f0d21dc58827c535f0))

## [0.1.4](https://github.com/Otard95/ed-expedition/compare/v0.1.3...v0.1.4) (2025-12-20)


### Bug Fixes

* **nix:** add gdk-pixbuf and libsoup_3 dependencies ([f1b4996](https://github.com/Otard95/ed-expedition/commit/f1b499657358a6537b6b25b2e4a26b41ede4b990))

## [0.1.3](https://github.com/Otard95/ed-expedition/compare/v0.1.2...v0.1.3) (2025-12-20)


### Bug Fixes

* **ci:** use Ubuntu 24.04 for webkit2_41 build ([fe965f6](https://github.com/Otard95/ed-expedition/commit/fe965f67429feb34218ff0ccb2793dbd00ef9aeb))
* **docs:** add release workflow documentation ([7044026](https://github.com/Otard95/ed-expedition/commit/704402661c7081c596a97ddf5de1bed2d8c2b17e))
* **nix:** use webkit2_41 variant and add missing glib dependency ([43018e6](https://github.com/Otard95/ed-expedition/commit/43018e69893b5a4ba54890227bd91ed8811acaa9))

## [0.1.2](https://github.com/Otard95/ed-expedition/compare/v0.1.1...v0.1.2) (2025-12-20)


### Bug Fixes

* **ci:** Should allow release to run ([4233bf3](https://github.com/Otard95/ed-expedition/commit/4233bf36a8055e6ad8eac7719719ab9d0f294a8f))

## [0.1.1](https://github.com/Otard95/ed-expedition/compare/v0.1.0...v0.1.1) (2025-12-20)


### Bug Fixes

* **docs:** Add commit section ([1acb8db](https://github.com/Otard95/ed-expedition/commit/1acb8db35e842cc1a2c56aef2e56d474d7b92043))

## [0.1.0](https://github.com/Otard95/ed-expedition/compare/v0.0.1...v0.1.0) (2025-12-20)


### Features

* **active:** Add completion modal with detailed stats ([cab8df9](https://github.com/Otard95/ed-expedition/commit/cab8df9993b3e90e90aece3f447314ac12fe9267))
* **api:** Add expedition management and plotter API ([40b024e](https://github.com/Otard95/ed-expedition/commit/40b024e6fd7799180a6ad68176814e9c707026e0))
* **app:** Add PlotRoute method for expedition route creation ([ccf0b53](https://github.com/Otard95/ed-expedition/commit/ccf0b5303b548207fc7cb91c6a91de5020f0198a))
* **app:** Add target event ([64f9807](https://github.com/Otard95/ed-expedition/commit/64f98079dfc2322d8885ff181db94966e5878e54))
* **app:** Emit events on jump history ([b89e9bf](https://github.com/Otard95/ed-expedition/commit/b89e9bf42ceebf9d0595f02e6031b6355c3663ad))
* **cmd:** Add prune-spansh-data utility ([4d3b76e](https://github.com/Otard95/ed-expedition/commit/4d3b76e051e34fc5bc322c3a38c8f110c1d875c4))
* **components:** Add compact mode to Table component ([42a1c5d](https://github.com/Otard95/ed-expedition/commit/42a1c5d72fe737a7abc897dd08f33ab62256a783))
* **components:** Add generic UI components ([d096021](https://github.com/Otard95/ed-expedition/commit/d0960219267b8b328f46932e310d507fee791f5e))
* **components:** Add reusable ConfirmDialog component ([8562f5d](https://github.com/Otard95/ed-expedition/commit/8562f5dd78ffdd0054ee6a23e637bd1f925af020))
* **components:** Add reusable icon components ([a75bd52](https://github.com/Otard95/ed-expedition/commit/a75bd52bbda5808857d743365df83bdb9a71fbc9))
* **components:** Add up and down direction to arrow ([d0615a7](https://github.com/Otard95/ed-expedition/commit/d0615a7cfbb7a0b2d1271e172a1ba82c93aee7ec))
* **database:** Add ED_EXPEDITION_DATA_DIR environment variable override ([2fe8d2e](https://github.com/Otard95/ed-expedition/commit/2fe8d2eed7239dcb7748def8ca44871e6dfbc204))
* **expedition-service:** Add fanout channel for jump history notifications ([526f16a](https://github.com/Otard95/ed-expedition/commit/526f16ad280f3bf1f234d64c9edf9e92f21cc1ed))
* **expedition:** Add completion event fanout channel ([525ba24](https://github.com/Otard95/ed-expedition/commit/525ba24b51b86d5e2ada1618ddd9ddf599bca6b6))
* **expedition:** Allow first jump to confirm starting location ([8ad9201](https://github.com/Otard95/ed-expedition/commit/8ad9201802a8afc013cd4043aaaf5dee579006a2))
* **expedition:** Implement FSDJump processing and auto-completion ([db775a9](https://github.com/Otard95/ed-expedition/commit/db775a916f408a925d321c15be7521974613062c))
* **expeditions:** Add modal-based delete confirmation ([78e91e3](https://github.com/Otard95/ed-expedition/commit/78e91e37fabea7b1a2f1ab77d2b196417d3edcf0))
* **expeditions:** Add route removal from expeditions ([6c10f53](https://github.com/Otard95/ed-expedition/commit/6c10f53de1d7c76d97635edda86ee615675ffa0a))
* **frontend:** Add hash-based routing with svelte-spa-router ([2ba198e](https://github.com/Otard95/ed-expedition/commit/2ba198e79f9786ad4415739a6a304046ea83bdd0))
* Implement expedition start transition (planned → active) ([1d5ee11](https://github.com/Otard95/ed-expedition/commit/1d5ee1178ab648bc368d163d1be55cc37fe5c188))
* Initial Commit ([0bff019](https://github.com/Otard95/ed-expedition/commit/0bff0199342684b741e00908b65a5c3a6be8e09a))
* **journal:** Add Location event tracking ([2a55c43](https://github.com/Otard95/ed-expedition/commit/2a55c430d9c32cd2c9b6ab2b1d7446895cebb099))
* **journal:** Implement journal sync with trace logging ([ad4fc52](https://github.com/Otard95/ed-expedition/commit/ad4fc523eeef636d2cc32100e8543e737a4516a8))
* **lib:** Add Find and fs utilities ([2340343](https://github.com/Otard95/ed-expedition/commit/23403433b35e5336ddc77b098bb9b1d494fad3d2))
* **license:** Add GPLv2 license ([eba7f0d](https://github.com/Otard95/ed-expedition/commit/eba7f0d15801ffecbab11f72fec4a533c3f7e0ec))
* **links:** Add backend link creation with validation ([2b4ddb7](https://github.com/Otard95/ed-expedition/commit/2b4ddb726489df550567ad10b8078500d5686e64))
* **links:** Add link deletion with X button on badges ([5a8c17f](https://github.com/Otard95/ed-expedition/commit/5a8c17fe82a7eeec54fe79d04b347b70377efece))
* **links:** Add route graph traversal and cycle detection ([e425949](https://github.com/Otard95/ed-expedition/commit/e425949b1dc0b3d997d3b317246a28647875d8e8))
* **links:** Implement link creation UI with cycle warnings ([d507c62](https://github.com/Otard95/ed-expedition/commit/d507c627046899e4b1a456d834a41fa8a232e8e8))
* **main:** Initialize journal watcher at startup ([8b78485](https://github.com/Otard95/ed-expedition/commit/8b78485d3a1e5212b6878e21c14c9d0100e4fd6a))
* **models:** Add app state and enhance expedition index ([ee0b0c3](https://github.com/Otard95/ed-expedition/commit/ee0b0c39bd774f222fb1e2a8bfcf2ba5c79c57db))
* **models:** Add Location tracking and enhance JumpHistoryEntry ([f2e8c43](https://github.com/Otard95/ed-expedition/commit/f2e8c4395663465c075245a2adecf82868755150))
* **plotters:** Add spansh data module with go:embed ([1e8dc09](https://github.com/Otard95/ed-expedition/commit/1e8dc0902103b61d4c2d0d162f74038cd2022a34))
* **plotters:** Implement plotter interface system ([498cc7e](https://github.com/Otard95/ed-expedition/commit/498cc7e71abe1b6c58f393401f4286c4bf76f4ed))
* **plotters:** Implement Spansh Galaxy Plotter HTTP integration ([ccf57bb](https://github.com/Otard95/ed-expedition/commit/ccf57bbb478d7118cc549a89623fbdd746255aa3))
* **route-table:** Add fuel_in_tank with fuel used as indicator ([f8a133c](https://github.com/Otard95/ed-expedition/commit/f8a133c5a5f0564309018c0bed0991aab5f0b952))
* **routes:** Add copy system name to clipboard in route table ([9be9524](https://github.com/Otard95/ed-expedition/commit/9be95245523fa4bd91180c3a43b3bba31a721772))
* **routes:** Add link creation UI with candidate selection ([902f80f](https://github.com/Otard95/ed-expedition/commit/902f80f2be683e5c729f9f18156ba73368ee4658))
* **routes:** Add must_refuel indicator to scoopable display ([1ccd32c](https://github.com/Otard95/ed-expedition/commit/1ccd32c3d48a7e1069041b005b4f64824e5a3f71))
* **routes:** Add Overcharge column with lightning indicator ([8b2727d](https://github.com/Otard95/ed-expedition/commit/8b2727d2ca050015a37208a51ebdd2307d881b58))
* **routes:** Integrate PlotRoute API with complete flow ([4844f2c](https://github.com/Otard95/ed-expedition/commit/4844f2c0c706e5cee38135c0c24a644a22f969aa))
* **services:** Add expedition CRUD operations ([b23842f](https://github.com/Otard95/ed-expedition/commit/b23842f3b3fd59fc56d84d1ae92580962e4b36e5))
* **services:** Track player location from FSDJump events ([c377614](https://github.com/Otard95/ed-expedition/commit/c3776140b746b64eec2756cd90e21112e85db9a4))
* **slice:** Add Map helper function ([9cc0226](https://github.com/Otard95/ed-expedition/commit/9cc02267de1d1596ee7760b7f220083d7aa41be1))
* **target:** Hook up frontend target indicator to event ([eced446](https://github.com/Otard95/ed-expedition/commit/eced4468d740420b89495ff78d3e67fafa9e4a57))
* **ui:** Add /active route and improve expedition navigation ([e3eaa20](https://github.com/Otard95/ed-expedition/commit/e3eaa20ea6bb22343da15635accfeec7042ecda6))
* **ui:** Add class prop passthrough to Button component ([a1c71ad](https://github.com/Otard95/ed-expedition/commit/a1c71ad9131e336122b86b137fe1269bc549dd6b))
* **ui:** Add route plotting wizard with dynamic input components ([8081094](https://github.com/Otard95/ed-expedition/commit/80810947120a9ae7526d8a264e8a8f513c542c59))
* **ui:** Implement active expedition view ([a865f75](https://github.com/Otard95/ed-expedition/commit/a865f7584c79b322b282a586594b05c9bbddbc0f))
* **ui:** Improve dropdown positioning and behavior ([535cd3b](https://github.com/Otard95/ed-expedition/commit/535cd3bb4d27386de8e86907275c713fe1946d4d))
* **ui:** Persist route collapse state across reloads ([a602540](https://github.com/Otard95/ed-expedition/commit/a602540a7423575d31d838eb59b76a8295461cd4))
* **view:** Add expedition view page for completed expeditions ([fe9617c](https://github.com/Otard95/ed-expedition/commit/fe9617c2e640686087431dfc7d29d1d30b04cd96))
* **views:** Add expedition edit UI with routing ([dc88ccf](https://github.com/Otard95/ed-expedition/commit/dc88ccf90119d0f4aff5fffc755f7e4a5e4a312d))
* **views:** Add New Expedition button to index view ([6a803d7](https://github.com/Otard95/ed-expedition/commit/6a803d7633443ce2c218be9e555e5e980165921b))
* **views:** Integrate route wizard into expedition edit view ([87e74c8](https://github.com/Otard95/ed-expedition/commit/87e74c89e65ad3c8673f1f52b0b4cbd57622a664))
* **views:** Load expedition data and add inline rename ([735d614](https://github.com/Otard95/ed-expedition/commit/735d6146eb25b7e3ab8432c15b5e8f0f2d27809c))


### Bug Fixes

* **components:** Move SVG size from attributes to style ([90f0344](https://github.com/Otard95/ed-expedition/commit/90f0344778d05684a18886d1c4615e57c08daeda))
* **database:** Use UnixNano for temp file names to prevent collisions ([ac18035](https://github.com/Otard95/ed-expedition/commit/ac1803568d96b86b77b8652d99e9f98c0267f97d))
* **events:** Fix timestamp comparison, now ignored identical time ([1e419bd](https://github.com/Otard95/ed-expedition/commit/1e419bdc4a1ce260509520005c227fff542a2f54))
* **frontend:** Add loading view before app is ready ([a18c65e](https://github.com/Otard95/ed-expedition/commit/a18c65eca5f52236c5789a2a72219494095be4bf))
* **frontend:** correct view button route for completed expeditions ([db82d21](https://github.com/Otard95/ed-expedition/commit/db82d217ab23e4f4a51fc4d43324a5e82ef742f5))
* **icons:** Use inline styles instead of SVG width/height attributes ([c55932c](https://github.com/Otard95/ed-expedition/commit/c55932ca788ecaba2c8e4b639cf3774a16454b80))
* **journal:** Fix buffer append in watcher file reading ([1a3c63d](https://github.com/Otard95/ed-expedition/commit/1a3c63da73402e0ec97a234df40b8d546e2a05b6))
* **journal:** Fix FanoutChannel timeout - was 200ns not 200ms ([76b0d58](https://github.com/Otard95/ed-expedition/commit/76b0d588ff30cdb58f78065ae4daf4392bee475f))
* **plotters:** Correct Spansh poll completion state check ([d8d4c20](https://github.com/Otard95/ed-expedition/commit/d8d4c2060d82dc1aab0fa90cfea2af17b01c4170))
* **plotters:** Handle Spansh API 2xx responses and float types ([199dde8](https://github.com/Otard95/ed-expedition/commit/199dde888add2125f5f039715a27f4f7be4df558))
* **services:** Add nil check for LastKnownLoadout to prevent SIGSEGV ([d8c7d19](https://github.com/Otard95/ed-expedition/commit/d8c7d19971ae1213c302f26f8d8f84008ba945e5))
* Spansh galaxy plotter and gen types ([6251aba](https://github.com/Otard95/ed-expedition/commit/6251abaa0b460508de620d9105aa75967acc469c))
* **state:** App state created/updated correctly ([bc0783a](https://github.com/Otard95/ed-expedition/commit/bc0783aa87a5fd444866aee31bd54c025d642017))
* **wails:** Restore build dir ([b2f74dc](https://github.com/Otard95/ed-expedition/commit/b2f74dce6bf978230ab69eda52b0aa81a48755dc))


### Performance Improvements

* **links:** Pre-compute link candidates in parent component ([04e207a](https://github.com/Otard95/ed-expedition/commit/04e207a21443182896f15c0b6becbacb214e5e07))
