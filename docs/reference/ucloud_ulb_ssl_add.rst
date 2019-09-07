.. _ucloud_ulb_ssl_add:

ucloud ulb ssl add
------------------

Add SSL Certificate

Synopsis
~~~~~~~~


Add SSL Certificate

::

  ucloud ulb ssl add [flags]

Options
~~~~~~~

::

  --region     string                  Optional. Override default region, see 'ucloud region'
                                       (default "cn-bj2") 

  --project-id     string              Optional. Override default project-id, see 'ucloud
                                       project list' (default "org-ryrmms") 

  --name     string                    Required. Name of ssl certificate to add 

  --format     string                  Optional. Format of ssl certificate (default "Pem") 

  --all-in-one-file     string         Optional. Path of file which contain the complete
                                       content of the SSL certificate, including the content
                                       of site certificate, the private key which encrypted
                                       the site certificate, and the CA certificate.  

  --site-certificate-file     string   Optional. Path of user's certificate file, *.crt.
                                       Required if all-in-one-file is omitted 

  --private-key-file     string        Optional. Path of private key file, *.key. Required if
                                       all-in-one-file is omitted 

  --ca-certificate-file     string     Optional. Path of CA certificate file, *.crt 

  --help, -h                           help for add 


Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

  --debug, -d   Running in debug mode 

  --json, -j    Print result in JSON format whenever possible 


Available Commands
~~~~~~~~

* :ref:`ucloud ulb ssl <ucloud_ulb_ssl>` 	 - List and manipulate SSL Certificates for ULB

