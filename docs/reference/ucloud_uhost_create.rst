.. _ucloud_uhost_create:

ucloud uhost create
-------------------

Create UHost instance

Synopsis
~~~~~~~~


Create UHost instance

::

  ucloud uhost create [flags]

Options
~~~~~~~

::

  --cpu     int                          Required. The count of CPU cores. Optional
                                         parameters: {1, 2, 4, 8, 12, 16, 24, 32} (default 4) 

  --memory-gb     int                    Required. Memory size. Unit: GB. Range: [1, 128],
                                         multiple of 2 (default 8) 

  --password     string                  Required. Password of the uhost user(root/ubuntu) 

  --image-id     string                  Required. The ID of image. see 'ucloud image list' 

  --async                                Optional. Do not wait for the long-running operation
                                         to finish. 

  --count     int                        Optional. Number of uhost to create. (default 1) 

  --vpc-id     string                    Optional. VPC ID. This field is required under
                                         VPC2.0. See 'ucloud vpc list' 

  --subnet-id     string                 Optional. Subnet ID. This field is required under
                                         VPC2.0. See 'ucloud subnet list' 

  --name     string                      Optional. UHost instance name (default "UHost") 

  --bind-eip     strings                 Optional. Resource ID or IP Address of eip that will
                                         be bound to the new created uhost 

  --create-eip-line     string           Optional. BGP for regions in the chinese mainland and
                                         International for overseas regions 

  --create-eip-bandwidth-mb     int      Optional. Required if you want to create new EIP.
                                         Bandwidth(Unit:Mbps).The range of value related to
                                         network charge mode. By traffic [1, 300]; by
                                         bandwidth [1,800] (Unit: Mbps); it could be 0 if the
                                         eip belong to the shared bandwidth 

  --create-eip-traffic-mode     string   Optional. 'Traffic','Bandwidth' or 'ShareBandwidth'
                                         (default "Bandwidth") 

  --shared-bw-id     string              Optional. Resource ID of shared bandwidth. It takes
                                         effect when create-eip-traffic-mode is ShareBandwidth  

  --create-eip-name     string           Optional. Name of created eip to bind with the uhost 

  --create-eip-remark     string         Optional.Remark of your EIP. 

  --charge-type     string               Optional.'Year',pay yearly;'Month',pay
                                         monthly;'Dynamic', pay hourly (default "Month") 

  --quantity     int                     Optional. The duration of the instance. N
                                         years/months. (default 1) 

  --project-id     string                Optional. Override default project-id, see 'ucloud
                                         project list' (default "org-ryrmms") 

  --region     string                    Optional. Override default region, see 'ucloud
                                         region' (default "cn-bj2") 

  --zone     string                      Optional. Override default availability zone, see
                                         'ucloud region' (default "cn-bj2-02") 

  --type     string                      Optional. Accept values: N1, N2, N3, G1, G2, G3, I1,
                                         I2, C1. Forward to
                                         https://docs.ucloud.cn/api/uhost-api/uhost_type for
                                         details (default "N2") 

  --net-capability     string            Optional. Default is 'Normal', also support 'Super'
                                         which will enhance multiple times network capability
                                         as before (default "Normal") 

  --os-disk-type     string              Optional. Enumeration value. 'LOCAL_NORMAL', Ordinary
                                         local disk; 'CLOUD_NORMAL', Ordinary cloud disk;
                                         'LOCAL_SSD',local ssd disk; 'CLOUD_SSD',cloud ssd
                                         disk; 'EXCLUSIVE_LOCAL_DISK',big data. The disk only
                                         supports a limited combination. (default "LOCAL_NORMAL") 

  --os-disk-size-gb     int              Optional. Default 20G. Windows should be bigger than
                                         40G Unit GB (default 20) 

  --os-disk-backup-type     string       Optional. Enumeration value, 'NONE' or 'DATAARK'.
                                         DataArk supports real-time backup, which can restore
                                         the disk back to any moment within the last 12 hours.
                                         (Normal Local Disk and Normal Cloud Disk Only)
                                         (default "NONE") 

  --data-disk-type     string            Optional. Enumeration value. 'LOCAL_NORMAL', Ordinary
                                         local disk; 'CLOUD_NORMAL', Ordinary cloud disk;
                                         'LOCAL_SSD',local ssd disk; 'CLOUD_SSD',cloud ssd
                                         disk; 'EXCLUSIVE_LOCAL_DISK',big data. The disk only
                                         supports a limited combination. (default "LOCAL_NORMAL") 

  --data-disk-size-gb     int            Optional. Disk size. Unit GB (default 20) 

  --data-disk-backup-type     string     Optional. Enumeration value, 'NONE' or 'DATAARK'.
                                         DataArk supports real-time backup, which can restore
                                         the disk back to any moment within the last 12 hours.
                                         (Normal Local Disk and Normal Cloud Disk Only)
                                         (default "NONE") 

  --firewall-id     string               Optional. Firewall Id, default: Web recommended
                                         firewall. see 'ucloud firewall list'. 

  --group     string                     Optional. Business group (default "Default") 

  --help, -h                             help for create 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud uhost <ucloud_uhost>` 	 - List,create,delete,stop,restart,poweroff or resize UHost instance

