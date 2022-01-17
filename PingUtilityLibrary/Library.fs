namespace PingUtilityLibrary

open System.Net
open System.Net.NetworkInformation
open System
open System.IO
open System.Threading
open System.Globalization
open System.Threading.Tasks


module PingUtility =
    let logFileName = "log.txt"
    let defaultProcessingLocation = "../../log"
    let configFileName = "config.txt"
    let googleUrl = "google.com"
    let defaultLocation = "Home-PC"
    let noRouter: IPAddress Option = None
    let optionaLocationGenerator ll = Some ll

    let mutable sw = File.AppendText(logFileName)
    
    type IPInfo = { IPAddress: IPAddress; Name: string }
    
    let pingAsync (target: IPAddress) =
        task {
            let pingSender = new Ping()
            let targetReply = pingSender.SendPingAsync(target, 250)
    
            let! result = targetReply
    
            return result
        }
        
    let pingAyncWithName target =
        task {
            try
                let! result = pingAsync target.IPAddress
                return target, Choice1Of2 result
            with 
            | ex -> 
                return target, Choice2Of2 ex
        }
    
    let pingAll (targets: IPInfo []) (router: IPInfo option) =
        let tasks = targets |> Array.map pingAyncWithName
        Task.WhenAll(tasks).Wait()
        tasks |> Array.mapi(fun i task -> 
            if task.IsCompletedSuccessfully then
                task.Result
            else
                targets[i], Choice2Of2(Exception("TimedOut"))
                )
        

    let pingToText info (reply: Choice<PingReply, exn>) =
        match reply with
        | Choice1Of2 pingResponse ->
            if pingResponse.Status = IPStatus.Success then
                sprintf "%s: %O ms" info.Name pingResponse.RoundtripTime, true
            else
                sprintf "%s: %O" info.Name pingResponse.Status, false
        | Choice2Of2 e -> sprintf "%s: %O" info.Name e.Message, false
    
    
    let ipInfoFromUrl (url: string) =
        let dnsInfo = Dns.GetHostEntry(url)
        let ip = dnsInfo.AddressList.[0]
        { IPAddress = ip; Name = url }
    
    
    let pingAndLog (targetUrl: string) routerIP location logLocation =
        let router = routerIP |> Option.map (fun r -> { IPAddress = r
                                                        Name = "router" })
    
        let target = ipInfoFromUrl targetUrl
        let google = ipInfoFromUrl googleUrl
    
        let targets = [| target; google |]
        let mutable oneHourFromLastWriteTime = DateTime.Now.AddHours(1)
    

        let local =
            { IPAddress = IPAddress.Loopback
              Name = "localhost" }
    
        let all =
            [| Some(local); router |] 
            |> Array.choose id
            |> Array.append targets

        let processingLocation = 
            match logLocation with
            | Some ll -> ll
            | None ->  match location with 
                       | l when l = defaultLocation -> defaultProcessingLocation
                       | _ -> $"{Directory.GetCurrentDirectory()}/old-logs/log"

        while true do
            let startTime = DateTime.Now
            let result =
                pingAll all router
                |> Array.map
                    (fun i ->
                        let info, reply = i
                        pingToText info reply)
    
            let timestamp =
                DateTime.Now.ToString("yyyy-MM-dd HH:mm:ss.fff", CultureInfo.InvariantCulture)
    
            let line = sprintf "%s: %s: %0A" timestamp location (result |> Array.map (fun (data, success) -> data))
            Console.WriteLine(line)
            if result |> Array.exists(fun (data, success) -> success = false) then
                sw.WriteLine(line)
                sw.Flush()
    
            if FileInfo(logFileName).Length > 5000000L || DateTime.Now > oneHourFromLastWriteTime then // 5 MB
                sw.Close()
                sw.Dispose()
    
                File.Move(logFileName, $"{processingLocation}-{DateTimeOffset.Now.ToUnixTimeSeconds().ToString()}.txt")
                File.Delete(logFileName)
                sw <- File.AppendText(logFileName)
                oneHourFromLastWriteTime <- DateTime.Now.AddHours(1.0)
            else
                ()
    
            let elapsedTime = (int)((DateTime.Now - startTime).TotalMilliseconds)
            Thread.Sleep(500 - Math.Min(elapsedTime, 250))
    
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
    
    let ``ask and ping`` routerIP location logLocation =
        let targetUrl = ``ask for target`` ()
        pingAndLog targetUrl routerIP location logLocation
    
