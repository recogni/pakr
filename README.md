# pakr

A golang UE4 `.pak` file reader.

## References

This code would have taken much much longer to whip up without the help of this repo: https://github.com/panzi/u4pak - much thanks.

## Usage

Currently this is somewhat WIP and only supports a subset of the pak functionality.  There is a `lib` exposed at `github.com/recogni/pakr/pak`.

`pak_list.go` implements a reader which parses a `.pak` file specified as the first command line argument to the program, and generates a list of all the assets that are contained in the file (along with their mount point).

```
go run pak_list.go <path/to/file.pak>
```
