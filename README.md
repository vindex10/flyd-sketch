# FlyD-sketch

Requirements:
* Requires root
* Requires aws cli in PATH

To prepare environment, you can use `./run.sh init`, which will:

* create storage files
* allocate thin pool
* create a symlink for imagecache

To run the server you can use: `go run ./src` or `./run.sh run`.

Communication goes via socket `./state/flyd-sketch.sock`,
this was also part of being careful with the env - not exposing the network socket.

To start an image, use:

```
curl -XGET --unix-socket ./state/flyd-sketch.sock -d 'python' start
```
