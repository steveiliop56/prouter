# Prouter

Prouter is an extremely simple Go program (>300 lines) that allows you to map directories with static assets to subdomains. This can be useful for when you want to share your markdown notes in a web form or want to serve your HTML files for testing.

## Running

You can run prouter by downloading the binary from the [latest](https://github.com/steveiliop56/prouter/releases/latest) release and running it as so:

```bash
./prouter --serve public --address 0.0.0.0 --port 8080
```

> [!NOTE]
> Make sure the binary is executable with `chmod +x prouter`.

Then you can visit `myapp.127.0.0.1.sslip.io:8080` and prouter will serve the contents of the `public/myapp` directory. You can either have raw HTML files or markdown files that will be rendered to HTML.

You can also run prouter with docker using the following docker run command:

```bash
docker run --rm --name prouter -v ./public:/public -p 8080:8080 ghcr.io/steveiliop56/prouter
```

There is also an available [docker compose](./docker-compose.yml) file.

## Building

You can build the app by cloning the repository:

```bash
git clone https://github.com/steveiliop56/prouter
```

Fetching the Go dependencies:

```bash
go mod download
```

Finally build the app with:

```bash
go build -o prouter main.go
```

## License

Prouter is licensed under the MIT License. TL;DR — You can use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the software. Just make sure to include the original license in any substantial portions of the code. There’s no warranty — use at your own risk.
See the [LICENSE](./LICENSE) file for full details.

## Contributing

If you like you can contribute by creating an issue or a pull request. All forms of contribution are welcome.

## Acknowledgements

This project is inspired by [Smallweb](https://github.com/pomdtr/smallweb) which is a lot more flexible and allows for bigger and more complex projects.
