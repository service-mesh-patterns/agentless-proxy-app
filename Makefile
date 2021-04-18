build_docker:
	docker build -t nicholasjackson/connect-native:v0.0.3 .

push_docker: build_docker
	docker push nicholasjackson/connect-native:v0.0.3