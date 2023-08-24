workspace {
    model {
        !include https://raw.githubusercontent.com/ministryofjustice/opg-technical-guidance/main/dsl/poas/persons.dsl
                !include https://raw.githubusercontent.com/ministryofjustice/opg-modernising-lpa/main/docs/architecture/dsl/local/makeRegisterSoftwareSystem.dsl
        !include lpaDeedService.dsl
        lpaCaseManagement = softwareSystem "LPA Case Management" "PKA Sirius." "Existing System" {
            -> apiGateway "Gets deeds from and sends updates to"
        }

        ualpa_SoftwareSystem = softwareSystem "Use A Lasting Power of Attorney" "Allows LPA Actors to retrieve and share LPAs with People and Organisations interested in LPAs" "Existing System" {
            -> apiGateway "Gets deeds from"
        }

        makeRegisterSoftwareSystem -> apiGateway "Sends deeds to"
    }

    views {
        systemContext lpaDeedService "SystemContext" {
            include *
            autoLayout
        }

        container lpaDeedService {
            include *
            autoLayout
        }

        theme default

        styles {
            element "Existing System" {
                background #999999
                color #ffffff
            }
            element "Web Browser" {
                shape WebBrowser
            }
            element "Database" {
                shape Cylinder
            }
        }
    }
}
