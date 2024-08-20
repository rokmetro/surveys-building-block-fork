# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Add CORS support
### Fixed
- Fix survey completed field [#69](https://github.com/rokwire/surveys-building-block/issues/69)
- Fix GET surveys time filtering

## [1.10.0] - 2024-07-30
### Added 
- Add admin GET survey responses API [#66](https://github.com/rokwire/surveys-building-block/issues/66)
### Added
- Sort Order when showing all (public) surveys [#64](https://github.com/rokwire/surveys-building-block/issues/64)

## [1.9.0] - 2024-07-26
### Added 
- Add "complete" field to show if the survey is completed [#61](https://github.com/rokwire/surveys-building-block/issues/61)
## [1.8.2] - 2024-07-26
### Fixed
- Fix Get /surveys [#58](https://github.com/rokwire/surveys-building-block/issues/58)

## [1.8.1] - 2024-07-26
### Fixed
- Fix "public" and "archived" [#55](https://github.com/rokwire/surveys-building-block/issues/55)

## [1.8.0] - 2024-07-25
### Added
- Fix "start_date" and "end_date" timestamp, "archived", "public" and set "complete" in the result [#52](https://github.com/rokwire/surveys-building-block/issues/52)

## [1.7.0] - 2024-07-24
### Added
- Add Estimated_completion_time [#49](https://github.com/rokwire/surveys-building-block/issues/49)

## [1.6.0] - 2024-07-23
### Added
- Archived flag [#40](https://github.com/rokwire/surveys-building-block/issues/40)
### Added
- Public flag [#37](https://github.com/rokwire/surveys-building-block/issues/37)
### Added
- Add extras field to survey data [#39](https://github.com/rokwire/surveys-building-block/issues/39)
### Fixed
- Fix "start_date" and "end_date" [#43](https://github.com/rokwire/surveys-building-block/issues/43)

## [1.5.0] - 2024-07-22
### Added 
- Start/end date [#38](https://github.com/rokwire/surveys-building-block/issues/38)

## [1.4.0] - 2024-07-11
### Added
- Remove user data [#34](https://github.com/rokwire/surveys-building-block/issues/34)

## [1.3.0] - 2023-09-20
### Added
- Reintroduce survey responses admin API [#27](https://github.com/rokwire/surveys-building-block/issues/27)

## [1.2.3] - 2023-09-19
### Fixed
- Fix delete survey APIs [#25](https://github.com/rokwire/surveys-building-block/issues/25)

## [1.2.2] - 2023-09-19
### Fixed
- Event admins cannot update or delete event surveys [#23](https://github.com/rokwire/surveys-building-block/issues/23)

## [1.2.1] - 2023-08-14
### Fixed
- Event admins cannot update surveys [#21](https://github.com/rokwire/surveys-building-block/issues/21)

## [1.2.0] - 2023-07-27
### Added
- Associate surveys with Calendar BB events [#19](https://github.com/rokwire/surveys-building-block/issues/19)
- Support survey creation tool [#14](https://github.com/rokwire/surveys-building-block/issues/14)

## [1.1.0] - 2023-03-15
### Added
- Create survey response analytics API [#11](https://github.com/rokwire/surveys-building-block/issues/11)
### Fixed
- Fix PerformTransaction [#9](https://github.com/rokwire/surveys-building-block/issues/9)

## [1.0.2] - 2022-12-09
### Fixed
- Survey response ID not set on create [#4](https://github.com/rokwire/surveys-building-block/issues/4)

## [1.0.1] - 2022-12-09
### Fixed
- Client APIs failing scope authorization [#1](https://github.com/rokwire/surveys-building-block/issues/1)

## [1.0.0] - 2022-12-09
### Added
- Initial implementation

[Unreleased]: https://github.com/rokwire/core-building-block/compare/v1.3.0...HEAD
[1.3.0]: https://github.com/rokwire/core-building-block/compare/v1.2.3...v1.3.0
[1.2.3]: https://github.com/rokwire/core-building-block/compare/v1.2.2...v1.2.3
[1.2.2]: https://github.com/rokwire/core-building-block/compare/v1.2.1...v1.2.2
[1.2.1]: https://github.com/rokwire/core-building-block/compare/v1.2.0...v1.2.1
[1.2.0]: https://github.com/rokwire/core-building-block/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/rokwire/core-building-block/compare/v1.0.2...v1.1.0
[1.0.2]: https://github.com/rokwire/core-building-block/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/rokwire/core-building-block/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/rokwire/core-auth-library-go/tree/v1.0.0
