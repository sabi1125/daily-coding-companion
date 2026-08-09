**Daily coding problem state diagram**

## Auth/Session state diagram
```mermaid
stateDiagram-v2
[*] --> SignedIn : sign in with google oauth
SignedIn --> SignedOut : actively sign out
SignedIn --> NeedsReAuth : signed out after 7 days <br> due to google specifications
NeedsReAuth --> SignedOut : actively sign out
NeedsReAuth --> SignedIn : reconnected
SignedOut --> SignedIn : sign in again
```

## Daily problem state diagram
```mermaid
stateDiagram-v2
[*] --> Open : problem created
Open --> Failed : submits, self-reports not solved
Open --> Solved : submits, self-reports solved
Failed --> Solved : resubmits, self-reports solved
Solved --> [*]
```

"Never touched" isn't a state — it's a derived view (an `Open` problem with zero submission
rows). Nothing transitions a problem there or out of it; it just is, until a real submission
event happens.

## Ingest Failure
```mermaid
stateDiagram-v2
[*] --> AttemptIngest : attempts to get daily coding problem
AttemptIngest --> Success : successfully gets problem <br> and populates database
AttemptIngest --> FailedToIngest : failed to get daily <br> coding problems 
FailedToIngest --> RetryIngest : retry ingest on login in
RetryIngest --> Success : successfully gets problem <br> and populates database
RetryIngest --> [*] : if retry also fails <br> determine it is not obtainable
Success --> [*]
```
