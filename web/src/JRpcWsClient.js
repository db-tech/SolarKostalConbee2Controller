import {Client} from "rpc-websockets";


export class JRpcWsClient {
    constructor(options = {}) {
        this.url = options.url || 'ws://localhost:8888/ws';
        this.isConnectedCallback = options.isConnectedCallback || (() => {
        });
        this.onOpenCallback = options.onOpen || (() => {
        });
        this.onMessageCallback = options.onMessage || (() => {
        });
        this.onCloseCallback = options.onClose || (() => {
        });
        this.onErrorCallback = options.onError || (() => {
        });
        this.name = options.name || this.url;
        this.isConnected = false;
        this.callbackMap = new Map();
        this.connect();
    }

    connect() {
        this.client = new Client(this.url, {max_reconnects: 0}, method => {
            //return a random string hash as the id
            return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
        });
        // this.client.subscribe('getRooms')
        this.client.on('open', () => {
            this.onOpenCallback()
            this.isConnected = true;
            this.isConnectedCallback(true)
        });
        this.client.on('message', (message) => {
            this.onMessageCallback(message)
            console.log("Received Notification: " + JSON.stringify(message));
        });
        this.client.on('close', () => {
            this.onCloseCallback()
            this.isConnected = false;
            this.isConnectedCallback(false)
        });
        this.client.on('error', (error) => {
            this.onErrorCallback(error)
            this.isConnected = false;
            this.isConnectedCallback(false)
        });
    }

    subscribe(method, callback) {
        // this.callbackMap.set(method,callback);
        console.log("Subscribing to " + method);
        // this.client.subscribe(method)
        this.client.on(method, callback)
    }

    callSimple(remoteFunction, params, onErrorCallback, onSuccessCallback) {
        this.client.call(remoteFunction, params).then((response) => {
            console.log(response);
            if (onSuccessCallback) {
                onSuccessCallback(response);
            }
            console.log("Response: " + response);
        }).catch((error) => {
                if (onErrorCallback) {
                    onErrorCallback(error);
                }
                console.error(error);
            }
        );
    }

    login(username) {
        this.callSimple("login", {"username": username})
    }

    async status() {
        return await this.client.call("status", {})
    }

    async authenticate(username, password, hostAddress) {
        return await this.client.call("authenticate", {
            "username": username,
            "password": password,
            "hostAddress": hostAddress
        })
    }

    async loginKostal(username, password, hostAddress) {
        return await this.client.call("loginKostal", {
            "username": username,
            "password": password,
            "hostAddress": hostAddress
        })
    }

    async getLights() {
        return await this.client.call("getLights", {})
    }

    async getProperties() {
        return await this.client.call("getProperties", {})
    }

    async saveProperties(properties) {
        return await this.client.call("saveProperties", properties)
    }

    stopMonitoring() {
        this.callSimple("stopMonitoring", {})
    }

    startMonitoring() {
        this.callSimple("startMonitoring", {})
    }

    async switchLightOn(lightId) {
        return await this.client.call("switchLightOn", {"lightId": lightId})
    }

    async switchLightOff(lightId) {
        return await this.client.call("switchLightOff", {"lightId": lightId})
    }

    init() {
        this.callSimple("init", {})
    }
}
