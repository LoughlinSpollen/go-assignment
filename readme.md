# Running the assignment client and server

## Server side

To run the server and RabbitMQ, use `make run` and afterwards `make stop`

### `make run` and `make stop`

Builds and runs the project.
This command can also be run from within the client and service sub-projects.

When run from the root it also installs and builds the client and service binaries. It then runs RabbitMQ and the service.

`make stop` will kill the service and stop the RabbitMQ.

**Usage:**
```sh
make run
make stop
```

## Client side

To run the client side, first build the client with `make build` and then run the binary in the `client\build` directory.

**Usage:**
```sh
make build
```

### CLI Commands

The CLI comands to send requests via RabbitMQ to the server are:

#### assignment_client add [KEY] [VALUE]
Add a key-value entry.

**Usage:**
```sh
cd client
./build/assignment_client add a b
```

#### assignment_client delete [KEY]
Delete an entry using a key.

**Usage:**
```sh
cd client
./build/assignment_client delete a
```

#### assignment_client get [KEY]
Get one item by key.

**Usage:**
```sh
cd client
./build/assignment_client get a
```

#### assignment_client getall
Get all key-value pairs.

**Usage:**
```sh
cd client
./build/assignment_client getall
```


# Make Commands

### `make install-dev`
Installs all development dependencies required for the project. This assumes you're working on MacOS. You can check the dependencies in `scripts/install-dev.sh`.

**Usage:**
```sh
make install-dev
```

### `make start-infra`
Starts the infrastructure (RabbitMQ message broker).

**Usage:**
```sh
make start-infra
```

### `make install`
Generates the mod files and downloads the dependency packages. The `mod` files should be generated with the script as it replaces the `shared_lib` package import directive.

**Usage:**
```sh
make install
```

### `make build`
Compiles the project and creates an two binary executables in the `client/build` and the `service/build` directories.
This command can also be run from within the client and service sub-projects.

**Usage:**
```sh
make build
```

### `make test`
Runs the unit the tests for the project.
This command can also be run from within the client, service and shared_lib sub-projects.

**Usage:**
```sh
make test
```

