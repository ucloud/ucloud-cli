## 0.1.39 (2022-06-13)

* fix init failure when user donot have a default project in ucloud console.
* update the description of init.

## 0.1.38 (2022-04-27)

* fix build failure when using go 1.18 on darwin_arm64. ([golang/go#49219](https://github.com/golang/go/issues/49219))

## 0.1.36 (2021-07-07)

ENHANCEMENTS:

* add the flags: `--instance-type`, `--forward-region`, `--bandwidth-package` about command `gssh create` to customize specific instance type of global ssh.
* add response fields: `GlobalSSHPort`, `InstanceType` about command `gssh list` to list specific instance type of global ssh.
* update cmd `uhost create` about bind EIP: EIP creates and binds to UHost when UHost creating instead of after UHost creating.

## 0.1.35 (2020-11-11)

ENHANCEMENTS:

* add the flag `--user-data-base64` about command `ucloud uhost create` to customize the startup behaviors when launching the uhost instance and the value must be base64-encode.(#55)

## 0.1.34 (2020-11-11)

ENHANCEMENTS:

* add the flag `--user-data` about command `ucloud uhost create` to customize the startup behaviors when launching the uhost instance.(#54)
* add the flag `--gpu-type` about command `ucloud uhost create` to define the type of GPU instance.(#54)

## 0.1.33

* Add command 'ucloud api', which can call any API of ucloud like this
    - ucloud api --Action DescribeUHostInstance --Region cn-bj2 or
    - ucloud api --local-file ./create_uhost.json
* Adapt to cloudshell

## 0.1.32

* Fixbug for creating uhost with shared bandwith. Now you can create uhost bound with shared bandwith using follow command.
```
ucloud uhost create --cpu 1 --memory-gb 2 --image-id uimage-xxx --password xxxxx --create-eip-traffic-mode ShareBandwidth --shared-bw-id bwshare-lxxxx
```

## 0.1.31

* fixbug, password missed when creating redis

## 0.1.30

* support creating uhost without data disk.
* default value of flag '--machine-type' changed to 'N' from empty when creating uhost.

## 0.1.29

* resize attached disk without stop uhost
* make batch creating uhost faster

## 0.1.28

* command 'ucloud uhost resize' add flag '--data-disk-id', to resize the specified udisk.
* fixbug #45

## 0.1.27

* Enable hot-plug for uhost when running 'ucloud uhost create'
* Add command 'ucloud uhost leave-isolation-group', 'ucloud uhost isolation-group create' and 'ucloud uhost isolation-group delete'

## 0.1.26

* fixbug about base-url

## 0.1.25

* ask permission for upload log when executing 'ucloud init'

## 0.1.24

* add global flags --base-url, --timeout-sec, --max-retry-times
* command [ucloud uhost create] add flag --hot-plug, --isolation-group
* add command [ucloud uhost isolation-group list]

## 0.1.23

* fix dead lock when creating uhosts in parallel
* refactor part of eip and ulb operations

## 0.1.22

* Add global flag '--public-key' and '--private-key' to override public-key and private-key in local config files.
* Add flag '--max-retry-times' for command 'ucloud config' so that users can set retry times for failed idempotent API calls.
* Add flag '--region-all' and '--output' for command 'ucloud uhost list' so that users can list uhosts in all regions and display more infomations about uhost.

## 0.1.21

* Add global flag '--profile' to specify profile for any command.
* Add command 'ucloud ext uhost switch-eip'

## 0.1.20

* Add command:
  ucloud pathx uga create | delete | list | describe | add-port | delete-port
  ucloud pathx upath list

## 0.1.19

* Bugfix for running command ucloud init failed.

## 0.1.18

* Add following commands:
    - `ucloud config add`
    - `ucloud config update`
    - `ucloud redis restart`  
    - `ucloud memcache restart`

* Command [ucloud uhost list --uhost-id-only] list uhost-ids  separated by comma
* Command [ucloud uhost delete --uhost-id xx,xx] can delete uhost instances concurrently.
  You can use [ucloud uhost delete --uhost-id \`ucloud uhost list --uhost-id-only --page-off\`] to delete all uhost instances in parallel.

## 0.1.17

* add flags page-off and uhost-id-only for uhost list

## 0.1.16

* Support log rotation. Log file path $HOME/.ucloud/cli.log.
* Bugfix for display nothing when uhost create failed

## 0.1.15

* Update documents 
* Add test for uhost

## 0.1.14

* Create uhost concurrently

## 0.1.13

* Update version of ucloud-sdk-go to fix bug

## 0.1.12

* Preliminary support umem
 
## 0.1.11

* Use go modules to manage dependencies
* Fix bug for uhost clone

## 0.1.10

* Support udb mysql

## 0.1.9

* Better flag value completion with local cache and multiple resource ID completion
* Command structure adjustment
  - ucloud bw-pkg => ucloud bw pkg
  - ucloud shared-bw => ucloud bw shared
  - ucloud ulb-vserver => ucloud ulb vserver
  - ucloud ulb-ssl-certificate => ucloud ulb ssl
  - ucloud ulb-vserver add-node/update-node/delete-node/list-node => ucloud ulb vserver backend add/update/delete/list
  - ucloud ulb-vserver add-policy/list-policy/update-policy/delete-policy => ucloud ulb vserver policy add/list/update/delete

## 0.1.8

* Support ulb

## 0.1.7

* Add udpn, firewall, shared bandwidth and bandwidth package; Refactor vpc, subnet and eip

## 0.1.6

* Improve uhost，image and disk-snapshot

## 0.1.5

* support batch operation.

## 0.1.4

* Support udisk.
* Polling udisk and uhost long time operation
* Async complete resource-id

## 0.1.3

* Integrate auto completion.
* Support uhost create, stop, delete and so on.

## 0.1.2

* Simplify config and completion.

## 0.1.1

* UHost list; EIP list,delete and allocate; GlobalSSH list,delete,modify and create.
