# Photlog Sample App in Go

This is just an exercise which builds something like imgur.com but in a more
basic form of course, but none the less this can be taken as a basis if you're
interested in building alike applications in go.

## How to run

After getting `go` installed, you simply clone the repo, put some initial images
(optional) in the `uploads/` folder and run it like so:

```sh
$ go install
$ photolog
```

This creates a server configured to run on `http://localhost:9000`.

Theoretically you can use `go get` to do all of this for yourself:

```sh
$ go get git@github.com:ZzAntares/photolog.git
$ photolog
```

## [LICENSE](https://www.gnu.org/licenses/gpl-3.0.txt)
