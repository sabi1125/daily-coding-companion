# Daily coding companion Architectural diagram

```mermaid
flowchart TD
    Visitor([Visitor, browser]) -->|HTTP| Website
    Gmail[[Gmail]]
    ClaudeAPI[[ClaudeAPI]]
    subgraph CloudHost
        Backend[Backend Server]
        Website[Web server]
        Cron[Cron]
        Database[(Database)]
    end
    Backend --> |Google SDK| Gmail
    Backend --> |GORM| Database
    Backend --> |Claude SDK| ClaudeAPI
    Website --> |AXIOS / REST| Backend
    Cron --> |Triggers Daily| Backend
```

> [!NOTE] OAuth consent happens directly between the browser and Google, not shown — this diagram covers steady-state API calls only.
