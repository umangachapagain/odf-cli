# Set

The set command supports the following sub-commands:

* [recovery-profile](#recovery-profile)
* [ceph-log-level](#ceph-log-level)

## recovery-profile

This command will set the recovery profile to favor either the new client IO, recovery IO, or a balanced mode. The default is `balanced` mode.
We can use the following built-in profile types:

* balanced
* high_client_ops
* high_recovery_ops

```bash
odf set recovery-profile high_client_ops
```

To verify the recovery profile setting run [odf get recovery-profile](get.md#recovery-profile).

## ceph-log-level

This command will set the log level for different ceph [subsystems](https://docs.ceph.com/en/latest/rados/troubleshooting/log-and-debug/#ceph-subsystems).
The `debug_` prefix will be automatically added to the subsystem when enabling the logging.

``` bash
odf set ceph-log-level osd crush 20
```

Once the logging efforts are complete, restore the systems to their default or to a level suitable for normal operations.