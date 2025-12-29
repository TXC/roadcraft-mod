This is a Go based repository with a Ruby client for certain API endpoints. It is primarily responsible for ingesting metered usage for GitHub and recording that usage. Please follow these guidelines when contributing:

This is a Go application that modifies different vehicles in the game RoadCraft from Sabre Interactive.
It reads a compressed file, named "default_other.pak" that is located at "root/paks/client/default" in the game directory inside the Steam Library.
The file is a actually a ZIP-file, compressed with "Store".

For example, on "Zikz 605E - Mobile Scalper", we need to update the file "ssl/autogen_designer_wizard/trucks/auto_ziks605e_mobile_scalper_res/auto_zikz_605e_mobile_scalper_res.cls" inside "default_other.pak".
- To allow scraping everywhere, we change the "allowedpercentage" to "0".
- To increase the quarry radius, "UsableCheckerDistance" from "150" to "600".

## Code Standards

### Required Before Each Commit
- Run `make fmt` before committing any changes to ensure proper code formatting
- This will run gofmt on all Go files to maintain consistent style

### Development Flow
- Build: `make build`
- Test: `make test`
- Full CI check: `make ci` (includes build, fmt, lint, test)

## Repository Structure
- `cmd/`: Main service entry points and executables
- `internal/`: Logic related to interactions with other GitHub services
- `docs/`: Documentation
- `testing/`: Test helpers and fixtures

## Key Guidelines
1. Follow Go best practices and idiomatic patterns
2. Use existing libraries and frameworks where possible
3. Maintain existing code structure and organization
4. Use dependency injection patterns where appropriate
5. Write unit tests for new functionality. Use table-driven unit tests when possible.
6. Document public APIs and complex logic. Suggest changes to the `docs/` folder when appropriate
7. Tests should be placed in the same package as the code they are testing, using `_test.go` suffix for test files.
8. Ensure all tests pass before submitting a pull request.
9. Try to achieve high test coverage for new code.
10. Use descriptive commit messages that explain the purpose of the changes.
11. Review existing code for style and consistency before adding new code.