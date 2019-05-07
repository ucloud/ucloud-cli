## Change Log
v0.1.16
* Support log rotation. Log file path $HOME/.ucloud/cli.log.
* Bugfix for display nothing when uhost create failed

v0.1.15
* Update documents 
* Add test for uhost

v0.1.14
* Create uhost concurrently

v0.1.13
* Update version of ucloud-sdk-go to fix bug

v0.1.12
* Preliminary support umem
 
v0.1.11
* Use go modules to manage dependencies
* Fix bug for uhost clone

v0.1.10
* Support udb mysql

v0.1.9
* Better flag value completion with local cache and multiple resource ID completion
* Command structure adjustment
  - ucloud bw-pkg => ucloud bw pkg
  - ucloud shared-bw => ucloud bw shared
  - ucloud ulb-vserver => ucloud ulb vserver
  - ucloud ulb-ssl-certificate => ucloud ulb ssl
  - ucloud ulb-vserver add-node/update-node/delete-node/list-node => ucloud ulb vserver backend add/update/delete/list
  - ucloud ulb-vserver add-policy/list-policy/update-policy/delete-policy => ucloud ulb vserver policy add/list/update/delete

v0.1.8
* Support ulb

v0.1.7
* Add udpn, firewall, shared bandwidth and bandwidth package; Refactor vpc, subnet and eip

v0.1.6
* Improve uhost，image and disk-snapshot

v0.1.5
* support batch operation.

v0.1.4
* Support udisk.
* Polling udisk and uhost long time operation
* Async complete resource-id

v0.1.3
* Integrate auto completion.
* Support uhost create, stop, delete and so on.

v0.1.2
* Simplify config and completion.

v0.1.1
* UHost list; EIP list,delete and allocate; GlobalSSH list,delete,modify and create.