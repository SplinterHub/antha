# antha
[![GoDoc](http://godoc.org/github.com/antha-lang/antha?status.svg)](http://godoc.org/github.com/antha-lang/antha)

This is a private fork of [antha](https://github.com/antha-lang/antha).

Antha v0.0.2

## OS X Installation Instructions

(See the primary antha repository for installation on other operating systems.)

1. Install [git](https://help.github.com/articles/set-up-git/) and configure it to use ssh with Github.
1. Install go: [golang.org](http://golang.org/doc/install).
1. If you don't have [Homebrew](http://brew.sh/), install it.
1. `brew install homebrew/science/glpk sqlite3`

We are following the Go workspace organization guidelines. Let's say `/Users/gmendel/Documents/splinter/` is where you are keeping all of your code, documentation, and results. The structure of that directory should eventually look like:

```
splinter/
  bin/
    (executables automatically get put here)
  pkg/
    (packages go in here)
  src/
    (write and put source code in here)
    github.com/
      SplinterHub/
        antha/
          (fork of antha)
        calculator/
          (imaginary splinter repository)
        visualizer/
          (imaginary splinter repository)
```
(This is not completely Go-Kosher, since you may write Go code for other projects than splinter, and would thus want a higher-level workspace directory.)

The following will create the `antha` directory under `SplinterHub`:

```
cd /Users/gmendel/Documents/splinter/

mkdir -p src/github.com/SplinterHub/

cd src/github.com/SplinterHub/
git clone git@github.com:SplinterHub/antha.git
```

In your `.profile` (or `.bash_profile`), write:

```
export GOPATH=/Users/gmendel/Documents/splinter/
export PATH=$PATH:$GOPATH/bin
export PATH=$PATH:/usr/local/go/bin
```

## Install antha

```
go get github.com/SplinterHub/antha/cmd/...
```

After following the installation instructions for your machine. You can check
if Antha is working properly by running a test protocol

```
cd $GOPATH/src/github.com/SplinterHub/antha/antha/examples/workflows/constructassembly

antharun --workflow workflow.json --parameters parameters.yml
```

## Making and Running Antha Components

The easiest way to start developing your own antha components is to place them
in the ``antha/component/an`` directory and follow the structure of the
existing components there. Afterwards, you can compile and use your components
with the following commands:

```
cd $GOPATH/src/github.com/antha-lang/antha
make clean && make
go get github.com/antha-lang/antha/cmd/...
antharun --workflow myworkflowdefinition.json --parameters myparameters.yml
```