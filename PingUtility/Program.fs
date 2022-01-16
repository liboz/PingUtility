open System
open System.IO

open PingUtilityLibrary


[<EntryPoint>]
let main argv =
    let location = match argv.Length with 
                   | 0 -> PingUtility.defaultLocation
                   | _ -> argv[0]
    Console.WriteLine($"Using location {location}")
    let routerIP = PingUtility.``find router ip`` ()

    match routerIP with
    | None when location = PingUtility.defaultLocation ->
        printfn "No Internet Connection"
        0
    | _ ->
        if argv.Length = 2 then
            PingUtility.pingAndLog argv[1] routerIP location None
        elif File.Exists(PingUtility.configFileName) then
            let lines = File.ReadAllLines(PingUtility.configFileName)

            if lines.Length <> 1 then
                PingUtility.``ask and ping`` routerIP location None
            else
                printfn "Use stored config file url? (Y/N)"
                let response = Console.ReadLine()

                if response = "Y" then
                    let targetUrl = lines |> Array.exactlyOne
                    PingUtility.pingAndLog targetUrl routerIP location None
                else
                    PingUtility.``ask and ping`` routerIP location None
        else
            PingUtility.``ask and ping`` routerIP location None

        0
