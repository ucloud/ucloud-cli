.. _ucloud_ulb_vserver_update:

ucloud ulb vserver update
-------------------------

Update attributes of VServer instances

Synopsis
~~~~~~~~


Update attributes of VServer instances

::

  ucloud ulb vserver update [flags]

Options
~~~~~~~

::

  --ulb-id     string                  Required. Resource ID of ULB instance which the VServer
                                       to create belongs to 

  --vserver-id     strings             Required. Resource ID of Vserver to update 

  --region     string                  Optional. Override default region, see 'ucloud region'
                                       (default "cn-bj2") 

  --project-id     string              Optional. Override default project-id, see 'ucloud
                                       project list' (default "org-ryrmms") 

  --name     string                    Optional. Name of VServer 

  --lb-method     string               Optional. LB methods, accept
                                       values:Roundrobin,Source,ConsistentHash,SourcePort,ConsistentHashPort,WeightRoundrobin and Leastconn. 
                                       ConsistentHash,SourcePort and ConsistentHashPort are effective for listen type PacketsTransmit only;
                                       Leastconn is effective for listen type RequestProxy only;
                                       Roundrobin,Source and WeightRoundrobin are effective for both listen types 

  --session-maintain-mode     string   Optional. The method of maintaining user's session.
                                       Accept values: 'None','ServerInsert' and 'UserDefined'.
                                       'None' meaning don't maintain user's session';
                                       'ServerInsert' meaning auto create session key;
                                       'UserDefined' meaning specify session key which
                                       accpeted by flag seesion-maintain-key by yourself 

  --session-maintain-key     string    Optional. Specify a key for maintaining session 

  --client-timeout-seconds     int     Optional.Unit seconds. For 'RequestProxy', it's
                                       lifetime for idle connections, range (0，86400]. For
                                       'PacketsTransmit', it's the duration of the connection
                                       is maintained, range [60，900] (default -1) 

  --health-check-mode     string       Optional. Method of checking real server's status of
                                       health. Accept values:'Port','Path' 

  --health-check-domain     string     Optional. Skip this flag if health-check-mode is
                                       assigned Port 

  --health-check-path     string       Optional. Skip this flags if health-check-mode is
                                       assigned Port 

  --help, -h                           help for update 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud ulb vserver <ucloud_ulb_vserver>` 	 - List and manipulate ULB Vserver instances

