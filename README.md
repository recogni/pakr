# pakr

A golang UE4 `.pak` file reader.

## References

This code would have taken much much longer to whip up without the help of this repo: https://github.com/panzi/u4pak - much thanks.

## Caution

Currently this "reader" is implemented with a single `main.go` and is not (yet) exposed as a library.  This was written to try and edit / trim `.pak` files created by the UnrealEngine.

## Run

```
go run main.go <path/to/file.pak>
```
