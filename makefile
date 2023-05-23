all: clean install

install:
	cd src/frontend && go build -buildvcs=false .
	cd src/orphan_reaper && go build -buildvcs=false .
	cd src/shim && go build -buildvcs=false .

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
	cp src/shim/shim /bin/podkit_shim

	rm src/frontend/frontend
	rm src/orphan_reaper/orphan_reaper
	rm src/shim/shim

clean:
	rm -f /bin/podkit
	rm -f /bin/podkit_orphan_reaper
	rm -f /bin/podkit_shim
	rm -rf /var/lib/podkit
