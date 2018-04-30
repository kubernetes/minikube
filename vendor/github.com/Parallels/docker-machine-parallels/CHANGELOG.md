# Changelog

## 1.3.0 (September 25, 2017)
- Fix compatibility with Parallels Desktop 13.0.* [GH-68]
- Add `--parallels-video-size` argument to manage a video memory size [GH-61]
- Use Golang 1.9 to build the binary [GH-69]
- Update `libmachine` libs to v0.12.2 [GH-69]
- Update `docker` libs to v17.03.2-ce [GH-69]
- Add missing dependencies to `vendor/` [GH-69]

## 1.2.3 (April 25, 2017)
- Update `libmachine` libs to v0.10.0

## 1.2.2 (January 10, 2017)
- Use Golang 1.7 to build the binary, update dependencies

## 1.2.1 (July 12, 2016)
- Force to connect the attached ISO image
- Check whether the host is connected to Shared network interface [GH-56]
- Update build dependencies [GH-57]

## 1.2.0 (April 15, 2016)
- Enable time sync feature by default [GH-47]
- Use Golang 1.6 to build the binary, update dependencies [GH-46]
- Update Boot2Docker cache in `PreCreateCheck` [GH-48]

## 1.1.1 (December 29, 2015)
- Compatibility with Docker Machine v0.5.5+ [GH-33]
- Add arguments `--version` and `--help` for plugin binary [GH-34]

## 1.1.0 (November 20, 2015)
- Compatibility with Docker Machine v0.5.1+

## 1.0.2 (November 20, 2015)
- Fixed issue with relative path in `MACHINE_STORAGE_PATH` [GH-17]

## 1.0.1 (November 11, 2015)
- Add unofficial support of Parallels Desktop 10 [GH-10]
- Remove `--nested-virt` and `--pmu-virt` from default VM settings [GH-13]

## 1.0.0 (November 4, 2015)
Initial public release
- Add `--parallels-no-share` flag [GH-1]
- Add acceptance tests
- Add build scripts
