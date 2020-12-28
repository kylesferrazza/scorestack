Medium Architecture
===================

This architecture deploys seven hosts:

- `elasticsearch1`, `elasticsearch2`, `elasticsearch3`: Elasticsearch master-eligible data nodes
- `elasticsearch4`: An Elasticsearch coordinating-only node
- `kibana`: The Kibana server
- `logstash`: The Logstash server
- `nginx`: An nginx server that proxies public requests to Kibana, Logstash, and the Elasticsearch cluster. Also used as an SSH jump box.

GCP Deployment
--------------

The Medium architecture can be deployed to GCP using multiple VMs using Terraform and Ansible. Before running the Terraform for the GCP deployment, make sure to [set up Terraform with the Google Provider](https://www.terraform.io/docs/providers/google/guides/getting_started.html) and make sure `git` is installed on your system. Then, [clone the Scorestack repository](cloning.md) to the host you plan to use for deploying Scorestack. This can be your laptop or wherever you're most comfortable - Scorestack itself will not be running in this host!

> Please note that new GCP accounts and projects have a per-region quota on CPU usage. This deployment will consume the entire quota.

### Deploying

First, change into the GCP terraform directory.

```shell
cd scorestack/deployment/medium/gcp
```

Next, provide values for the five unset variables within `variables.tf`. Here is a brief description of the variables:

- `project`: The GCP project ID to which Scorestack will be deployed
- `credentials_file`: The path to the GCP credentials file - see the [GCP provider reference](https://www.terraform.io/docs/providers/google/guides/provider_reference.html#credentials) for more information
- `ssh_pub_key_file`, `ssh_priv_key_file`: The paths to the SSH keypair that will be added to the created instances - these must already be created
- `fqdn`: The domain name that Scorestack will be served behind - the DNS record for this domain must be configured manually after deployment

> If you don't have a domain for Scorestack and were planning on accessing it via its public IP, you can set `fqdn` to an empty string.

Finally, run Terraform to deploy Scorestack.

```shell
terraform init
terraform apply
```

Once deployment is finished, configure the DNS record for your FQDN to point at the public IP of the `nginx` instance. Then, run the [ansible deployment](#ansible-deployment) for the medium architecture.

### Cleanup/Teardown

To destroy the Scorestack cluster completely and remove all artifacts, you must destroy the Terraform resources, the generated certificates, and the Ansible inventory file.

```shell
terraform destroy
rm -rf ../ansible/certificates
rm ../ansible/inventory.ini
```

Ansible Deployment
------------------

Once the infrastructure for the medium architecture has been deployed by one of the available Terraform options, Ansible is used to deploy and configure Scorestack. You will need to install Ansible on your system for this deployment.

### Deploying

Once your Terraform deployment is complete, change into the Ansible directory in the repository.

```shell
# If you're in one of the medium architecture Terraform folders
cd ../ansible

# If you're at the root level of the repository
cd deployment/medium/ansible
```

Then, run the Ansible playbook to deploy Scorestack. The inventory file was generated for you by Terraform to properly configure SSH access for Ansible to all of the instances.

```shell
ansible-playbook playbook.yml -i inventory.ini
```

### Dynamicbeat Certificates

To deploy Dynamicbeat, you will need the Dynamicbeat client certificate and key and the Scorestack internal CA certificate. These certificates are generated by the Terraform deployment process. Once the deployment has finished, these files can be found at the following paths within the repository:

- `deployment/medium/ansible/certificates/dynamicbeat/dynamicbeat.key`
- `deployment/medium/ansible/certificates/dynamicbeat/dynamicbeat.crt`
- `deployment/medium/ansible/certificates/ca/ca.crt`

See the [Dynamicbeat deployment guide](dynamicbeat.md) for more information.