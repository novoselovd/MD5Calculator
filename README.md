# MD5 Calculator

A small tool that calculates md5 hash of files

## Getting Started

### Installing

- Clone this repo to your local machine.

- Install dependencies via Glide:

```
glide install
```

- Set *WORKING_DIR*(optional)

- Run the app

```
go run app.go
```

### Usage

```
curl -X POST -d "url=http://site.com/file.txt" http://127.0.0.1:5000/submit
```
```
curl -X GET http://127.0.0.1:5000/check?id=c4ca4238a0b923820dcc509a6f75849b
```

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

