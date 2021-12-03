name: Create a Bug (English)
description: Report an issue in English
body:
  - type: markdown
    id: problem
    attributes:
      label: What Happened?
      description: |
      Tip: Add the "--alsologtostderr" flag to the command-line for more logs
    validations:
      required: true
  - type: markdown
    id: logs
    attributes:
      description: |
      label: Attach log file
      Tip: Run `minikube logs --file=log.txt` then drag & drop `log.txt` file to the browser. 
    validations:
      required: true
  - type: dropdown
    id: operating-system
    attributes:
      label: Operating System
      description: What is your OS ?
      options:
        - macOS (Default)
        - Windows
        - Ubuntu
        - Redhat/Fedora
        - Other
    validations:
      required: false