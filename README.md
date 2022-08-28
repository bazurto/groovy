# groovy

## Usage

## Create a groovy release

```bash
    make release VERSION=4.0.4
    make clean VERSION=4.0.4
    make release VERSION=3.0.9
    make clean VERSION=3.0.9
```

or

```bash
    for i in `cat versions`; do make release VERSION=$i; make clean $i; done
```
