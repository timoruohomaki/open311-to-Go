# Open311-to-Go


## What is Open311?

Open311 is a form of technology that provides open channels of communication for issues that concern public space and public services. Primarily, Open311 refers to a standardized protocol for location-based collaborative issue-tracking. By offering free web API access to an existing 311 service, Open311 is an evolution of the phone-based 311 systems that many cities in North America offer.

Unlike the synchronous one-to-one communication of a 311 call center, Open311 technologies use the internet to enable these interactions to be asynchronous and many-to-many. This means that several different people can openly exchange information centered around a single public issue. This open model allows people to provide more actionable information for those who need it most and it encourages the public to be engaged with civic issues because they know their voices are being heard. Yet Open311 isn’t just about this more open internet-enabled model for 311 services, it’s also about making sure the technology itself is open so that 311 services and applications are interoperable and can be used everywhere.

Source: https://www.open311.org/learn/

## Why Go?

Most of the Open311 implementations so far are based on Python on Django framework. While easy to implement, that might not be the most optimal starting point for a highly performant and scalable API backend.

## Anything new here?

This implementation inludes the additions the City of Helsinki added on Open311, mainly support for other languages and locales and support for external media server in cases where images are included in the ticket.

Source: https://dev.hel.fi/apis/open311 

This implementation also uses MongoDB as a backend, also utilizing its spatial functions. XML formats are supported with schemas.

Due to the experimental nature of this implementation, the schema for service request is extended with inline properties object, containing user-annotated properties as key-value pairs. This approach makes it possible to support use cases where there are additional properties in the service request, e.g. because of supporting a specific standard such as the Finnish PSK 5970 that defines the schema for data record of cases and events. With this approach, the goal is to link citizen feedback with ISO 55000 asset management practises.

## Development framework and versions

* This work uses golang version 1.24.1 [^1]. The work depends on the new Go net/http routing capabilities so a version of 1.22 or newer is required.
* The API will eventually be deployed as an Azure Function because it will then be easier to transfer to production platform. The dev server is Ubuntu 22.04 hosted at api.spatialworks.fi
* The development is done using Visual Studio Code - however it shouldn't make any difference what editor to use
* Sentry is used for telemetry (free version will be enough).
* API examples and tests collection is managed with [Bruno](https://www.usebruno.com/)

## Implementation Status (initial implementation)

* [ ]  Github action for Ubuntu ci/cd pipeline
* [ ]  Apache Combined Log Format on access logs
* [ ]  Service Discovery (Serf)
* [ ]  Observability (Sentry)
* [ ]  MongoDB database backend
* [ ]  Security (TLS, authentication, authorization)
* [ ]  Schema validation on XML messages
* [ ]  GET Service List (xml and json)
* [ ]  GET Service Definition (xml and json)
* [ ]  POST Service Request (xml and json)
* [ ]  GET Service Request Id (xml and json)
* [ ]  GET Service Requests (xml and json)
* [ ]  GET Service Request (xml and json)

## What is the motivation for this?

This work is to support my master's thesis work on large scale asset management on urban digital twins. This should be up and running by the end of 2024-ish.

## Credits

* This work heavily relies on the concept of distributed services Travis Jeffery provided in his book [Distributed Services with Go](https://a.co/d/g5mhjd8).
* Credits also to Ishan Shrestha on RestAPI and MongoDB best practises, [blog here](https://medium.com/@ishan.shrestha356/scalable-json-restapi-using-go-lang-and-mongodb-cf9699c5f6e8)



[^1]: It should be always possible to continue development using the latest stable release

