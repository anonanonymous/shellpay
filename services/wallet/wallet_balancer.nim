import asynchttpserver, asyncdispatch, httpclient, json, os


#[ is_ready - checks if the wallet service is ready for requests ]#
proc is_ready(client: HttpClient, host="127.0.0.1", port, password: string): bool =
    let payload = %*{
        "jsonrpc": "2.0",
        "password": password,
        "method": "getStatus"
    }
    try:
        let resp = client.request("http://" & host & ":" & port & "/json_rpc",
                                 httpMethod = HttpPost,
                                 body = $payload)
        let body = parseJSON(resp.body)
        return body["result"]["knownBlockCount"].getInt() - body["result"]["localDaemonBlockCount"].getInt() < 2
    except:
        echo "timeout"
        return false

#[ handler - http handler for wallet service requests ]#
proc handler(req: Request) {.async.} =
    case req.url.path
    of "/json_rpc":
        let max_attempts = 10
        var i = 1
        var client = newHttpClient(timeout = 1000)
        client.headers = newHttpHeaders({"Content-Type": "application/json"})

        for _ in 0 .. max_attempts:
            echo paramStr(i)
            if client.is_ready(port = paramStr(i), password = getEnv("RPC_PASS")):
                break
            i = if i < paramCount(): i + 1 else: 1
            sleep(500)

        let headers = newHttpHeaders({"Location": "http://localhost:"&paramStr(i)&"/json_rpc"})
        await req.respond(Http307, "", headers)
    else:
        await req.respond(Http404, "")

#[ main - Entry point ]#
proc main() =
    if paramCount() < 1:
        echo "Usage: RPC_PASS=<password> ./wallet_balancer <list of wallet service ports>"
        quit()
 
    var server = newAsyncHttpServer()
    waitFor server.serve(Port(8069), handler)

main()