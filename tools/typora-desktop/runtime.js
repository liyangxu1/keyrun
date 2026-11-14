function (config) {
    var fs = require("fs");
    var crypto = require("crypto");
    var path = require("path");
    var url = require("url");
    var entry = path.resolve(__filename);
    var marker = Buffer.from(config.marker);
    var original = Buffer.from(config.original, "base64");

    // 仅匹配当前入口，其他文件读取照常执行。
    function isEntry(value) {
        try {
            var target = value instanceof URL ? url.fileURLToPath(value) : value;
            if (Buffer.isBuffer(target)) target = target.toString();
            if (typeof target !== "string") return false;
            var resolved = path.resolve(target);
            return process.platform === "win32"
                ? resolved.toLowerCase() === entry.toLowerCase()
                : resolved === entry;
        } catch (_) {
            return false;
        }
    }

    function originalSource(options) {
        var encoding = typeof options === "string" ? options : options && options.encoding;
        return encoding ? original.toString(encoding) : Buffer.from(original);
    }

    var readSync = fs.readFileSync;
    fs.readFileSync = function (file, options) {
        return isEntry(file) ? originalSource(options) : readSync.apply(this, arguments);
    };
    var read = fs.readFile;
    fs.readFile = function (file, options, callback) {
        if (!isEntry(file)) return read.apply(this, arguments);
        if (typeof options === "function") {
            callback = options;
            options = null;
        }
        if (typeof callback !== "function") return read.apply(this, arguments);
        process.nextTick(function () { callback(null, originalSource(options)); });
    };
    var readAsync = fs.promises.readFile;
    fs.promises.readFile = function (file, options) {
        return isEntry(file) ? Promise.resolve(originalSource(options)) : readAsync.apply(this, arguments);
    };

    var decrypt = crypto.publicDecrypt;
    crypto.publicDecrypt = function (key, input) {
        if (!Buffer.isBuffer(input) || !input.equals(marker)) return decrypt.apply(this, arguments);
        return Buffer.from(JSON.stringify({
            deviceId: "0MA-",
            fingerprint: config.fingerprint,
            email: "research@example.invalid",
            license: "KEYRUN-TYPORA-RESEAR-CHONLY",
            version: "win|" + config.version,
            date: Date.now(),
            type: ""
        }));
    };

    var net = require("electron").net;
    var request = net.request;
    net.request = function (options) {
        var destination;
        try {
            destination = new URL(typeof options === "string" ? options : options.url || options.href);
        } catch (_) {
            return request.apply(this, arguments);
        }
        if (destination.protocol !== "https:" ||
            !["store.typora.io", "dian.typora.com.cn"].includes(destination.hostname) ||
            destination.pathname !== "/api/client/renew") {
            return request.apply(this, arguments);
        }
        var client = new (require("events").EventEmitter)();
        var closed = false;
        var sent = false;
        var headers = Object.create(null);
        client.setHeader = function (name, value) { headers[name.toLowerCase()] = value; };
        client.getHeader = function (name) { return headers[name.toLowerCase()]; };
        client.removeHeader = function (name) { delete headers[name.toLowerCase()]; };
        client.write = function () { return !closed; };
        client.abort = client.destroy = function () { closed = true; };
        client.end = function () {
            if (sent || closed) return;
            sent = true;
            process.nextTick(function () {
                if (closed) return;
                var body = JSON.stringify({ success: true, code: 1, msg: marker.toString("base64") });
                var response = require("stream").Readable.from([Buffer.from(body)]);
                response.statusCode = 200;
                response.statusMessage = "OK";
                response.headers = { "content-type": "application/json", "content-length": String(Buffer.byteLength(body)) };
                client.emit("response", response);
            });
        };
        return client;
    };
}
