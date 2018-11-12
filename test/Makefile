init:
	cd terraform && terraform init
	furyctl install

apply:
	cd terraform && terraform apply && terraform output inventory > ../ansible/hosts.ini

destroy:
	cd terraform && terraform destroy

run-ansible:
	cd ansible && ansible-playbook setup.yml


test:
	cd .. && go build && cp furyagent test.sh furyagent.yml test/ansible/files/
	$(MAKE) apply
	$(MAKE) run-ansible

