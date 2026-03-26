# KANBAN 

This README is humanmade.


## Planning (Or showcase 😉)

```mermaid
kanban
  section To Do
    TASK-001@{ label: "💎 DORA-metrics trackers 💎" }
    TASK-002@{ label: "✨ CLI board templates ✨" }
    TASK-003@{ label: "✨ CLI card templates ✨" }
    TASK-004@{ label: "✨ Customisable card fields ✨" }
    TASK-008@{ label: "💎 Acceptance-test code rules 💎" }
    TASK-009@{ label: "💎 Update kanban-board as pre-commit hook 💎" }
    TASK-012@{ label: "remove kanban ci-completion" }
  section In Progress
    TASK-010@{ label: "go releases not downloading" }
  section Done
    TASK-005@{ label: "💎 Renovation bot 💎" }
    TASK-006@{ label: "💎 Dependency vulnerability scan 💎" }
    TASK-007@{ label: "💎 Pipeline bashscripts entirely in makefile 💎" }
    TASK-011@{ label: "✨ Remove add command, since new command exists" }
    TASK-013@{ label: "Generate kanban mermaid as pre-commit step" }
```





## Look ma, no hands!

This project spawned out of being an experiment on getting AI to generate high quality code. There is currently a high influx of poorly designed ai generated 'wibe-code' on the internet. According to [DORA](https://dora.dev/research/) teams that their research group already characterised as "High performing" are able to greatly benefit from gen-AI whereas other teams have not. This project attempts to take available information regarding routines, markers, habbits of high performing teams, and incorporate them and asks the question:

**Is it possible succesfully create a high quality project with minimal manual coding intervention?**

## Installation

**Install via Homebrew:**

```
brew tap jmsargent/kanban
brew install kanban
```

**Install via go install:**

```
go install github.com/jmsargent/Kanban/cmd/kanban@latest
```

Alternatively you can download a binary from the [releases page](https://github.com/jmsargent/Kanban/releases)
