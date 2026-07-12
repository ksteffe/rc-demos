FROM maven:3.9-eclipse-temurin-17 AS build

WORKDIR /src
COPY extensions ./extensions
COPY rc-demos/apps/request-logger-http-java/pom.xml ./rc-demos/apps/request-logger-http-java/pom.xml
COPY rc-demos/apps/request-logger-http-java/src ./rc-demos/apps/request-logger-http-java/src

WORKDIR /src/rc-demos/apps/request-logger-http-java
RUN mvn -q package

FROM eclipse-temurin:17-jre

COPY --from=build /src/rc-demos/apps/request-logger-http-java/target/classes /app

ENTRYPOINT ["java", "-cp", "/app", "io.runtimeconditions.demos.requestlogger.RequestLoggerHttp"]
