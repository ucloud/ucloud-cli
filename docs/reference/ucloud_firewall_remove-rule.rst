.. _ucloud_firewall_remove-rule:

ucloud firewall remove-rule
---------------------------

Remove rule from firewall instance

Synopsis
~~~~~~~~


Remove rule from firewall instance

::

  ucloud firewall remove-rule [flags]

Examples
~~~~~~~~

::

  ucloud firewall remove-rule --fw-id firewall-2cxxxz/test.lxj2 --rules "TCP|24|0.0.0.0/0|ACCEPT|HIGH" --rules-file firewall_rules.txt

Options
~~~~~~~

::

  --fw-id     strings       Required. Resource ID of firewalls to update 

  --rules     strings       Required if rules-file is empay. Rules to add to firewall.
                            Schema:'Protocol|Port|IP|Action|Level'. See 'ucloud firewall
                            create --help' for detail. 

  --rules-file     string   Required if rules is empty. Path of rules file, in which each rule
                            occupies one line. Schema: Protocol|Port|IP|Action|Level. 

  --region     string       Optional. Region, see 'ucloud region' (default "cn-bj2") 

  --project-id     string   Optional. Project-id, see 'ucloud project list' (default "org-ryrmms") 

  --help, -h                help for remove-rule 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud firewall <ucloud_firewall>` 	 - List and manipulate extranet firewall

