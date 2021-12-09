FROM mcr.microsoft.com/dotnet/sdk:6.0

COPY Database/Release/net6.0/publish/ App/
WORKDIR /App
RUN mkdir /App/old-logs
ENTRYPOINT ["dotnet", "PingUtility.dll", "DigitalOcean", "facebook.com"]