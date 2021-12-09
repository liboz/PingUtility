dotnet publish -c Release
docker build -t pingutility .
docker save -o container pingutility
