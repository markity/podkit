all: clean install

install:
	cd src/frontend && go build -buildvcs=false .
	cd src/orphan_reaper && gcc main.c -o orphan_reaper
	cd src/shim/start && go build -buildvcs=false -o podkit_shim .
	cd src/shim/exec_back && gcc main.c -o podkit_shim_exec_back
	cd src/shim/exec_front && gcc main.c -o podkit_shim_exec_front

	mkdir -p /var/lib/podkit
	mkdir -p /var/lib/podkit/container
	mkdir -p /var/lib/podkit/images
	mkdir -p /var/lib/podkit/socket

	touch /var/lib/podkit/lock
	cp install_files/running_info.json /var/lib/podkit
	cp install_files/images/image_info.json /var/lib/podkit/images
	cp install_files/images/ubuntu2204.tar /var/lib/podkit/images
	cp src/frontend/frontend /bin/podkit
	cp src/orphan_reaper/orphan_reaper /bin/podkit_orphan_reaper
	cp src/shim/start/podkit_shim /bin/podkit_shim
	cp src/shim/exec_back/podkit_shim_exec_back /bin/podkit_shim_exec_back
	cp src/shim/exec_front/podkit_shim_exec_front /bin/podkit_shim_exec_front


	rm src/frontend/frontend
	rm src/orphan_reaper/orphan_reaper
	rm src/shim/start/podkit_shim
	rm src/shim/exec_back/podkit_shim_exec_back
	rm src/shim/exec_front/podkit_shim_exec_front

clean:
	rm -f /bin/podkit
	rm -f /bin/podkit_orphan_reaper
	rm -f /bin/podkit_shim
	rm -f /bin/podkit_shim_exec_back
	rm -f /bin/podkit_shim_exec_front
	rm -rf /var/lib/podkit
