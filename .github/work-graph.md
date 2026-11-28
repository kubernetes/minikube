<!-- witful-work-graph:begin -->

```mermaid
---
config: {"flowchart": {"rankSpacing": 90, "subGraphTitleMargin": {"bottom": 24, "top": 0}}, "layout": "dagre", "theme": "base", "themeVariables": {"edgeLabelBackground": "#F8F9F4", "lineColor": "#647D8B", "primaryTextColor": "#243746"}}
---
flowchart TB
    subgraph n8d0c7ac992af2f4913b47d425af0fe08dca35d157cbf27e40c203589b983d1fd["`project: Minikube integration fork`"]
        n2ad562319767157087dda0dec6391f4479f8a04869ab0cc8d3a9c3637dae73b5["branch: fix/worker-registration-before-labels (active)"]
    end
    subgraph n5086431107cae0151fdd8a231177374dfd063cb427d03273af07d527a23c27d0["`project: Minikube`"]
        nfc613b4dfd6736a7bd268c8a0e74ed0d1c04a959f59dd74ef2874983fd443fc9["branch: master"]
    end
    n2ad562319767157087dda0dec6391f4479f8a04869ab0cc8d3a9c3637dae73b5 -->|based on| nfc613b4dfd6736a7bd268c8a0e74ed0d1c04a959f59dd74ef2874983fd443fc9
    click n2ad562319767157087dda0dec6391f4479f8a04869ab0cc8d3a9c3637dae73b5 href "https://github.com/astrivant/minikube/tree/fix/worker-registration-before-labels" "Open recorded work" _blank
    click nfc613b4dfd6736a7bd268c8a0e74ed0d1c04a959f59dd74ef2874983fd443fc9 href "https://github.com/kubernetes/minikube/tree/master" "Open recorded work" _blank
    classDef work fill:#F8F9F4,stroke:#647D8B,color:#243746
    class n2ad562319767157087dda0dec6391f4479f8a04869ab0cc8d3a9c3637dae73b5 work
    style n8d0c7ac992af2f4913b47d425af0fe08dca35d157cbf27e40c203589b983d1fd fill:#F8F9F4,stroke:#829B85,color:#243746
    class nfc613b4dfd6736a7bd268c8a0e74ed0d1c04a959f59dd74ef2874983fd443fc9 work
    style n5086431107cae0151fdd8a231177374dfd063cb427d03273af07d527a23c27d0 fill:#F8F9F4,stroke:#829B85,color:#243746
    linkStyle 0 stroke:#829B85,color:#243746
```

Containment shows ownership; arrows name recorded relationships.

Associated work:

- [branch: fix/worker\-registration\-before\-labels](<https://github.com/astrivant/minikube/tree/fix/worker-registration-before-labels>)
- [project: Minikube integration fork](<https://github.com/astrivant/minikube>)
- [branch: master](<https://github.com/kubernetes/minikube/tree/master>)
- [project: Minikube](<https://github.com/kubernetes/minikube>)

<!-- witful-work-graph:end -->
