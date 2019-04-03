.. _ucloud_config:

ucloud config
-------------

Configure UCloud CLI options

Synopsis
~~~~~~~~


Configure UCloud CLI options such as private-key,public-key,default region and default project-id.

::

  ucloud config [flags]

Examples
~~~~~~~~

::

  ucloud config --profile=test --region=cn-bj2 --active

Options
~~~~~~~

::

  --profile     string       Required. Set name of CLI profile 

  --public-key     string    Optional. Set public key 

  --private-key     string   Optional. Set private key 

  --region     string        Optional. Set default region. For instance 'cn-bj2' See 'ucloud
                             region' 

  --zone     string          Optional. Set default zone. For instance 'cn-bj2-02'. See 'ucloud
                             region' 

  --project-id     string    Optional. Set default project. For instance 'org-xxxxxx'. See
                             'ucloud project list 

  --base-url     string      Optional. Set default base url. For instance
                             'https://api.ucloud.cn/' (default "https://api.ucloud.cn/") 

  --timeout-sec     int      Optional. Set default timeout for requesting API. Unit: seconds
                             (default 15) 

  --active                   Optional. Mark the profile to be effective 

  --help, -h                 help for config 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud <ucloud>` 	 - UCloud CLI v0.1.14
* :ref:`ucloud config delete <ucloud_config_delete>` 	 - delete settings by profile name
* :ref:`ucloud config list <ucloud_config_list>` 	 - list all settings

