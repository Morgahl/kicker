Kicker is a simple service for managing misbehaving Kubernetes deployments.


Documentation is to come. In the meantime a short description.


This program aims to manage misbehaving Kubernetes Deployments. Namely when they are running into issues that require periodic killing of a pod based on defined fleet behavior. This killing is done naively so be VERY CAREFUL that you use a well thought out configuration.


Also be aware that multiple Criteria pointing at the same resources can potentially cause all Pods in a Deployment to be killed at the same time.


I am also debating that this be useable as a scheduler for Kubernetes. Strategies would have to be configurable for both modes or we'll need to pick one mode or the other. I favor both modes as it give you a quick production ready pod killer and a scheduler based solution (albeit with a bit more planning involved)