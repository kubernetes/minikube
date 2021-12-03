name: English
description: Report an issue in English
body:
  - type: markdown
    id: problem
    attributes:
      label: What Happend?
      description: |
      Tip: Add the "--alsologtostderr" flag to the command-line for more logs
    validations:
      required: true
  - type: markdown
    id: problem
    attributes:
      description: |
      label: Attach log file
      Tip: Run `minikube logs --file=log.txt`) then drag & drop `log.txt` file to the browser. 
    validations:
      required: true
  - type: dropdown
    id: Operating system
    attributes:
      label: Operating Syste,
      description: What is your OS ?
      options:
        - MacOs (Default)
        - Windows
        - Ubuntu
        - Redhat/Fedora
        - Other
    validations:
      required: false