# DEVOPS Decisions: kanban-web-view

**Date**: 2026-03-29
**Status**: Draft
**Wave**: DEVOPS

---

## DD-DEVOPS-01: No Terraform -- Manual VM Provisioning

**Decision**: The GCP VM is provisioned manually via `gcloud` CLI commands, not Terraform.

**Rationale**: One VM, one environment, one developer. Terraform adds a state management burden (remote backend, state locking) for zero reproducibility benefit when there is exactly one server. The provisioning commands are documented in `infrastructure.md` and are idempotent (`gcloud compute instances create` fails gracefully if the VM exists).

**Revisit when**: A second environment (staging) is needed, or the project moves to multiple VMs.

---

## DD-DEVOPS-02: SCP-Based Deployment (Not Container Registry)

**Decision**: The `kanban-web` binary is cross-compiled in CI and uploaded to the VM via `scp`. No Docker images, no container registry.

**Rationale**: A statically-linked Go binary is a single file. Docker adds image build time, registry cost (or self-hosting), and container runtime on the VM. For a single binary on a single VM, `scp` + `systemctl restart` is the simplest deployment path. The e2-micro's 1 GB RAM cannot comfortably run Docker alongside nginx and the application.

**Trade-off**: No image versioning or layer caching. Accepted because binary builds are fast (< 30 seconds) and the VM only needs the latest version.

**Revisit when**: Multiple services are deployed to the VM, or a container orchestrator is introduced.

---

## DD-DEVOPS-03: SSH Key in CircleCI Context (Not GCP IAM)

**Decision**: CI deploys via SSH using a key stored in the CircleCI `deploy-context`, not via GCP IAM service account with `gcloud compute scp`.

**Rationale**: `gcloud compute scp` requires installing the gcloud SDK in the CI image (300+ MB) and configuring a service account with Compute Engine access. A plain SSH key is lighter, faster, and has a smaller blast radius (access to one VM, not the entire GCP project). The deploy user on the VM has `sudoers` restricted to specific commands only.

**Security note**: The SSH key is stored as a base64-encoded environment variable in the CircleCI context. Rotate the key quarterly or on team changes.

---

## DD-DEVOPS-04: Stop Inactive Instance After Switch

**Decision**: After a blue/green deployment switch, the previously active instance is stopped (not left running).

**Rationale**: The e2-micro has 0.25 vCPU and 1 GB RAM. Running two instances of `kanban-web` plus nginx simultaneously degrades performance. Stopping the old instance frees resources for the active one. The trade-off is that rollback requires starting the old service (adds ~2 seconds), but this is acceptable for the project's availability target (95%).

---

## DD-DEVOPS-05: UptimeRobot + Cron (Not GCP Monitoring)

**Decision**: External monitoring uses UptimeRobot free tier. Internal monitoring uses a cron-based health check script.

**Rationale**: GCP Cloud Monitoring requires agent installation and configuration. UptimeRobot provides external HTTP monitoring with email alerts at zero cost and zero setup on the VM. The cron script handles auto-restart for transient failures. Together they cover the two monitoring needs: "is the service healthy?" (cron) and "is the VM reachable from the internet?" (UptimeRobot).

---

## DD-DEVOPS-06: No Separate Staging Environment

**Decision**: There is no staging environment. CI tests validate the build. Deployment goes directly to production (the single VM).

**Rationale**: A staging environment would require a second VM (not free tier -- only one e2-micro per project is free). The blue/green deployment pattern provides a safety net: the new version is health-checked on the inactive instance before traffic switches. If the health check fails, no traffic is routed to the new version.

**Risk mitigation**: Web E2E tests in CI exercise the HTTP handlers with a real git repo. The deploy health check validates the binary starts and serves requests. Rollback is < 30 seconds.

---

## DD-DEVOPS-07: Shared Git Clone Between Blue and Green

**Decision**: Both blue and green instances read from the same local git clone at `/var/kanban/repo`.

**Rationale**: The git clone contains task files that are the same for both instances. Duplicating the clone wastes disk space and creates sync complexity. The write path is serialized by a mutex in the use case layer, so concurrent writes from two instances are not possible (and only one instance serves traffic at a time).

**Risk**: During a brief window after traffic switches, if the new instance has a different version of `AddTaskAndPush`, it could write task files in an incompatible format. Mitigated by the fact that the task file format (YAML front matter + markdown) is defined in the domain layer, which both versions share.

---

## DD-DEVOPS-08: Self-Signed TLS Initially, Let's Encrypt When Domain Available

**Decision**: Start with self-signed TLS certificates. Switch to Let's Encrypt when a domain name is confirmed.

**Rationale**: Let's Encrypt requires a domain name. The user has not confirmed a domain yet. Self-signed certificates enable HTTPS immediately (required for `Secure` cookies), with browser warnings as the known trade-off. The nginx config is structured so switching to Let's Encrypt is a single `certbot` command.

---

## DD-DEVOPS-09: Deploy Script in cicd/ (Not Ansible/Chef)

**Decision**: The deployment script is a shell script at `cicd/deploy-web.sh`, not an Ansible playbook or configuration management tool.

**Rationale**: The deployment is a sequence of 5 steps: upload binary, restart service, health check, switch nginx, stop old service. A shell script expresses this directly. Ansible adds a Python dependency, inventory management, and playbook authoring overhead for a single-host deployment.

**Revisit when**: More than one server is deployed to, or the deployment steps become complex enough to benefit from idempotent task declarations.
