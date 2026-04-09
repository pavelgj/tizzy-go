# Contributing to Tizzy

Thank you for your interest in contributing to Tizzy!

## Visual Regression Testing

Tizzy uses visual regression testing to ensure that components render correctly and that changes do not inadvertently break the UI appearance.

### How it works

- Tests generate PNG snapshots of components in memory.
- These snapshots are compared against "golden" reference images stored in `tz/testdata/golden/`.
- If the generated image differs from the golden reference, the test fails.
- Failed tests save the new (incorrect) image to `tz/testdata/failed/` so you can inspect the difference.

### Running Tests

To run all tests, including visual regression tests:

```bash
go test ./tz -v
```

To run only the visual tests:

```bash
go test ./tz -v -run "TestGenerate.*Visual"
```

### Updating Golden References

If you intentionally change the visual appearance of a component, you must update the golden reference images. You can do this by running the tests with the `-update` flag (it is recommended to use `-run` to only target visual tests and avoid running unrelated failing tests):

```bash
go test ./tz -v -run "TestGenerate.*Visual" -update
```

Make sure to review the updated images in `tz/testdata/golden/` before committing them.
