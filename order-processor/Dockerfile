FROM maven:3.9-eclipse-temurin-21 AS build
WORKDIR /src
COPY pom.xml .
RUN mvn -q -DskipTests dependency:go-offline
COPY src src
RUN mvn -q package -DskipTests
FROM eclipse-temurin:21-jre
WORKDIR /app
COPY --from=build /src/target/order-processor-1.0.0.jar app.jar
ENTRYPOINT ["java","-jar","app.jar"]

