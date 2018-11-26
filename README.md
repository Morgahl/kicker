Kicker is a simple service for managing misbehaving Kubernetes deployments.


Documentation is to come. In the meantime a short description.


This program aims to manage misbehaving Kubernetes Deployments. Namely when they are running into issues that require periodic killing of a pod based on defined fleet behavior. This killing is done naively so be VERY CAREFUL that you use a well thought out configuration.


Also be aware that multiple Criteria pointing at the same resources can potentially cause all Pods in a Deployment to be killed at the same time.