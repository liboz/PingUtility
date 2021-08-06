open System.Net
open System.Net.NetworkInformation
open System
open System.IO
open System.Threading
open System.Globalization

let logFileName = "log.txt"
let logBackUpFileName = "backup.txt"
let configFileName = "config.txt"
let googleUrl = "google.com"

let mutable sw = File.AppendText(logFileName)

type IPInfo = { IPAddress: IPAddress; Name: string }

let pingAsync (target: IPAddress) =
    async {
        let pingSender = new Ping()
        let targetReply = pingSender.SendPingAsync(target, 250)

        let! result = Async.AwaitTask targetReply

        return result
    }

let pingAsyncCatch target = pingAsync target |> Async.Catch

let pingAyncWithName target =
    async {
        let! result = pingAsyncCatch target.IPAddress

        return target, result
    }

let pingAll (targets: IPInfo []) (router: IPInfo) =
    let local =
        { IPAddress = IPAddress.Loopback
          Name = "localhost" }

    let all =
        [| local; router |] |> Array.append targets

    all
    |> Array.map pingAyncWithName
    |> Async.Parallel
    |> Async.RunSynchronously

let pingToText info (reply: Choice<PingReply, exn>) =
    match reply with
    | Choice1Of2 pingResponse ->
        if pingResponse.Status = IPStatus.Success then
            sprintf "%s: %O ms" info.Name pingResponse.RoundtripTime
        else
            sprintf "%s: %O" info.Name pingResponse.Status
    | Choice2Of2 e -> sprintf "%s: %O" info.Name e.Message


let ipInfoFromUrl (url: string) =
    let dnsInfo = Dns.GetHostEntry(url)
    let ip = dnsInfo.AddressList.[0]
    { IPAddress = ip; Name = url }


let pingAndLog (targetUrl: string) routerIP =
    let router =
        { IPAddress = routerIP
          Name = "router" }

    let target = ipInfoFromUrl targetUrl
    let google = ipInfoFromUrl googleUrl

    let targets = [| target; google |]

    while true do
        let result =
            pingAll targets router
            |> Array.map
                (fun i ->
                    let info, reply = i
                    pingToText info reply)

        let timestamp =
            DateTime.Now.ToString("yyyy-MM-dd HH:mm:ss.fff", CultureInfo.InvariantCulture)

        let line = sprintf "%s: %A" timestamp result
        sw.WriteLine(line)
        Console.WriteLine(line)
        sw.Flush()

        if FileInfo(logFileName).Length > 5000000L then // 5 MB
            sw.Close()
            sw.Dispose()

            if File.Exists(logBackUpFileName) then
                File.Delete(logBackUpFileName)
            else
                ()

            File.Move(logFileName, logBackUpFileName)
            File.Delete(logFileName)
            sw <- File.AppendText(logFileName)
        else
            ()

        Thread.Sleep(500)

let ``ask for target`` () =
    printfn "Input target url: "
    let targetUrl = Console.ReadLine()
    printfn "Save target url? (Y/N)"
    let response = Console.ReadLine()

    if response = "Y" then
        use configFile = File.CreateText(configFileName)
        configFile.WriteLine(targetUrl)

    targetUrl

let ``find router ip`` () =
    let adapters =
        NetworkInterface.GetAllNetworkInterfaces()

    let routerAdapter =
        adapters
        |> Array.map (fun a -> a.GetIPProperties())
        |> Array.tryFind
            (fun ap ->
                let addresses = ap.DhcpServerAddresses
                addresses.Count > 0)

    routerAdapter
    |> Option.map (fun ap -> ap.DhcpServerAddresses.[0])

let ``ask and ping`` routerIP =
    let targetUrl = ``ask for target`` ()
    pingAndLog targetUrl routerIP

[<EntryPoint>]
let main argv =
    let routerIP = ``find router ip`` ()

    match routerIP with
    | None ->
        printfn "No Internet Connection"
        0
    | Some r ->
        if File.Exists(configFileName) then
            let lines = File.ReadAllLines(configFileName)

            if lines.Length <> 1 then
                ``ask and ping`` r
            else
                printfn "Use stored config file url? (Y/N)"
                let response = Console.ReadLine()

                if response = "Y" then
                    let targetUrl = lines |> Array.exactlyOne
                    pingAndLog targetUrl r
                else
                    ``ask and ping`` r
        else
            ``ask and ping`` r

        0
