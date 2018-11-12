init:
	cd terraform && terraform init
	furyctl install

apply:
	cd terraform && terraform apply -auto-approve && terraform output inventory > ../ansible/hosts.ini

destroy:
	cd terraform && terraform destroy

run-ansible:
	cd .. && go build && mv -f furyagent test/ansible/files/
	cd ansible && ansible-playbook setup.yml

test:
	$(MAKE) apply
	$(MAKE) run-ansible
