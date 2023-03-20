package spec

var oneofSpec = []byte(`components:
  schemas:
    Cat:
      type: object
      properties:
        name:
          type: string
        scratches:
          type: boolean
        $type:
          type: string
          enum:
            - cat
      required:
        - name
        - scratches
        - $type
    Dog:
      type: object
      properties:
        name:
          type: string
        barks:
          type: boolean
        $type:
          type: string
          enum:
            - dog
      required:
        - name
        - barks
        - $type
    Animal:
      type: object
      oneOf:
        - $ref: "#/components/schemas/Cat"
        - $ref: "#/components/schemas/Dog"
      discriminator:
        propertyName: $type
        mapping:
          cat: "#/components/schemas/Cat"
          dog: "#/components/schemas/Dog"
`)

var oneofNoDiscriminatorSpec = []byte(`components:
  schemas:
    Cat:
      type: object
      properties:
        name:
          type: string
        scratches:
          type: boolean
      required:
        - name
        - scratches
    Dog:
      type: object
      properties:
        name:
          type: string
        barks:
          type: boolean
      required:
        - name
        - barks
    Animal:
      type: object
      oneOf:
        - $ref: "#/components/schemas/Cat"
        - $ref: "#/components/schemas/Dog"
`)
