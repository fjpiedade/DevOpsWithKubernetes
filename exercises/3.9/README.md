# Exercise 3.9 – DBaaS vs DIY

## Scenario

This comparison considers two ways of running PostgreSQL for the Todo App project:

- **DBaaS:** PostgreSQL using Google Cloud SQL, a managed database service.
- **DIY:** PostgreSQL deployed and managed by ourselves in Kubernetes on Google Kubernetes Engine (GKE).

## 1. Initialization

| DBaaS – Cloud SQL | DIY – PostgreSQL on GKE |
|---|---|
| Create an instance in Google Cloud, select the PostgreSQL version, region, CPU, memory and storage. | Create Kubernetes manifests for PostgreSQL, persistent storage, services, credentials and configuration. |
| Configure network access, database users and application connectivity. | Configure PVCs, resource limits, health checks, secrets and database access. |
| No need to install or manage the database operating system. | Requires knowledge of PostgreSQL, Kubernetes storage and database administration. |

**DBaaS advantages:** Faster setup and less initial configuration. The provider manages the database infrastructure.

**DIY advantages:** Full control over the PostgreSQL configuration, deployment and environment.

**Conclusion:** Cloud SQL requires less initial work. DIY requires more preparation and testing before it is safe for production.

## 2. Costs

| DBaaS – Cloud SQL | DIY – PostgreSQL on GKE |
|---|---|
| Pay for the database instance, based on CPU, memory and usage time. | Pay for the GKE compute resources consumed by PostgreSQL. |
| Storage and retained backups generate additional costs. | Persistent disks and backup storage generate additional costs. |
| High availability and additional features can increase the price. | Replicas, additional nodes and more storage can increase infrastructure costs. |
| Reduces infrastructure administration, but database administration, access control and cost management are still necessary. | Requires time and expertise for installation, monitoring, upgrades, backups and recovery. |

**DBaaS advantages:** Costs are easier to associate with a managed database service, and less operational work is required.

**DIY advantages:** Can be cheaper for small workloads when an existing Kubernetes cluster has spare capacity.

**Conclusion:** DIY may have a lower infrastructure cost in an existing cluster, but the operational effort is a hidden cost. Cloud SQL generally costs more directly but reduces the amount of maintenance work.

Actual costs depend on the selected configuration, region, storage, availability requirements and usage. There is no universally cheaper option.

## 3. Maintenance

| DBaaS – Cloud SQL | DIY – PostgreSQL on GKE |
|---|---|
| Google manages the underlying infrastructure and many database maintenance tasks. | We are responsible for PostgreSQL installation, configuration and upgrades. |
| Managed monitoring and integration with Google Cloud observability tools. | We must configure monitoring, logging, alerts and resource management. |
| Infrastructure maintenance and supported database operations are handled by the provider. | We must plan maintenance, security updates and database recovery procedures. |
| High availability and automatic failover are available when configured. | High availability requires additional configuration, replication and failover procedures. |

**DBaaS advantages:** Less operational responsibility and simpler maintenance.

**DIY advantages:** More control over versions, extensions, configuration and maintenance schedules.

**Conclusion:** Cloud SQL is easier to maintain. DIY provides flexibility but requires continuous administration and operational knowledge.

## 4. Backups and Recovery

| DBaaS – Cloud SQL | DIY – PostgreSQL on GKE |
|---|---|
| Supports scheduled automated backups and on-demand backups. | Backups must be configured and scheduled by us. |
| Supports Point-in-Time Recovery (PITR), when configured, to recover data to a specific time. | PITR requires additional configuration, such as WAL archiving and a suitable backup tool. |
| Backups can be restored through Google Cloud Console or the `gcloud` CLI. | Recovery requires restoring the database using tools such as `pg_restore` or a backup solution such as pgBackRest. |
| Backup retention and storage are configurable and may incur additional costs. | We must manage backup retention, storage, security and monitoring. |

For DIY deployments, possible backup methods include:

- `pg_dump`: creates logical backups. Custom-format dumps can be restored with `pg_restore`; plain SQL dumps are restored with `psql`.
- pgBackRest: supports physical backups, WAL archiving and point-in-time recovery.
- GKE Persistent Disk snapshots: provide volume-level recovery. For a database-consistent backup, coordinate the snapshot with PostgreSQL or use a database-aware backup procedure; a raw volume snapshot alone should not be treated as a guaranteed consistent backup.

**DBaaS advantages:** Backups and recovery are integrated into the service and require less manual work.

**DIY advantages:** Complete control over backup schedules, tools, storage locations and retention policies.

**Conclusion:** Cloud SQL offers a simpler backup and recovery experience. DIY can provide equivalent flexibility, but requires careful configuration and regular restore testing.

## 5. Overall Comparison

| Requirement | Better fit |
|---|---|
| Fast initialization | DBaaS |
| Less operational maintenance | DBaaS |
| Integrated backup and recovery | DBaaS |
| Full configuration control | DIY |
| Custom PostgreSQL setup | DIY |
| Reusing spare Kubernetes capacity | DIY |
| Less database administration work | DBaaS |

## Final Conclusion

For a production application where reliability, recovery and reduced operational effort are important, **DBaaS (Google Cloud SQL)** is generally the better choice.

Google manages much of the database infrastructure and provides integrated backup, recovery and monitoring capabilities. This allows the development team to focus more on the application.

**DIY PostgreSQL on GKE** is a good option when full control is required, when the team has Kubernetes and database administration experience, or when an existing cluster has spare resources.

However, DIY transfers responsibility for upgrades, security, monitoring, backups and disaster recovery to the team.

For the Todo App project, Cloud SQL would be a sensible production choice. The DIY approach remains useful for learning Kubernetes, PostgreSQL administration and persistent storage.

## References

- [Google Cloud SQL for PostgreSQL](https://cloud.google.com/sql/postgres)
- [Cloud SQL pricing](https://cloud.google.com/sql/pricing)
- [Cloud SQL backups and recovery](https://cloud.google.com/sql/docs/postgres/backup-recovery/backups)
- [Cloud SQL Point-in-Time Recovery](https://cloud.google.com/sql/docs/postgres/backup-recovery/pitr)
- [GKE Persistent Disk volume snapshots](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/backup-pd-volume-snapshots)
- [PostgreSQL documentation – Backup and Restore](https://www.postgresql.org/docs/current/backup.html)
- [pgBackRest documentation](https://pgbackrest.org/)