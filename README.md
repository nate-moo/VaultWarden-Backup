## Preface
While this is called VaultWarden-Backup It can be used as a cronjob container for any folder to file compressor. \
For proper disclosure, I utilized Generative AI while creating this tool. As such please always take care when using Generative AI output as it is known to hallucinate non-existent libraries, old libraries, or even malicious ones.

## Why I did this
There doesn't seem to be a decent backup facility for Vaultwarden, which to my understanding expects an external backup solution via backing up the disk or FS. However since I am using a non-standard storage solution, I can only easily mount my vaultwarden data into a kubernetes pod.

## How to use
You probably shouldn't, I've quickly whipped this up without regarding proper practices including the use of an LLM. It was created as an exercise for myself and for learning about LLMs within coding applications.

To utilize, you can either manually run a cron job executing the container or use the facilities of your container orchestrator to periodically run this container. \
This container's behavior is to immediately exit after completing the job. If the container fails it will print into stdout a failure message.

Volumes:
| Use    | Path | desc. |
| -------- | ------- | ---- |
| Data  | /data | Data you want to back up, for me, it's my vaultwarden database and similar applicable files |
| Backup | /backups | The location you want the /data backup to reside. |

flags:
| Flag | Default | Description |
| --- | --- | --- |
| source | /data | Directory to be compressed |
| target | /backups | Directory where tarballs are placed |
| verbose | false | Determines verbosity, file addition logging |

ENV:
| Env Var | Description |
| ------- | ----------- |
| VWBSOURCE | Directory to be compressed |
| VWBTARGET | Directory where tarballs are placed |

Environment variables trump flags, flags trump defaults, defaults are `/data` and `/backups`

This container simply takes the `/data` folder/mount, tars it, compresses it using ZSTD, and outputs it into the provided `/backups` mount point.

example:
\# ```docker run -v ./data:/data -v /mnt/NAS/Backups:/backups ghcr.io/nate-moo/vaultwarden-backup:1.1```

### running without the container
You are able to run this program without a OCI compliant system, however you must modify the paths in the `main.go` file to fit the backup and data locations

