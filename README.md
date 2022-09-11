# groovy

## Usage

1. Download the latest version of Bazurto CLI from https://github.com/bazurto/bz
2. Create a `.bz` file in your root of your project with the following content:
    ```hcl
    deps = [
        "github.com/bazurto/groovy-v4.0.4"
    ]
    ```
3. Create a groovy script `myscript.groovy`:
    ```groovy
    println "Hello World"
    ```
4. Now you can execute groovy commands in your project. e.g.:
    ```bash
    bz groovy myscript.groovy
    ```

    or 


    ```bash
    bz bash
    groovy myscript.groovy
    ```

## Contributing to this project

### Create Groovy Releases

```bash
    make release VERSION=4.0.4
    make clean VERSION=4.0.4
    make release VERSION=3.0.9
    make clean VERSION=3.0.9
```

or

```bash
    for i in `cat versions`; do make clean VERSION=$i; make release VERSION=$i; make clean VERSION=$i; done
```
