This library is responsible for handling all of the routing.
See godoc.org/github.com/AutoRoute/node for the full documentation

And see https://docs.google.com/document/d/1NBk83bfj6MLgDD6USidSethKqV6-cNa-ZO4odNrdcCc/edit?usp=sharing for a high level description of the project.

See https://docs.google.com/presentation/d/1b_Gl22d4e5oD5Z_4RMf-gjaCdrmqAzev6fSIaG7-uHw/edit?usp=sharing for a presentation about the project.

If you'd like to just spin up a node and start playing around with it, you can either run build and run them locally, or just use docker.

```
sudo docker run -p 30000:34321 --name p1  c00w/autoroute:latest -fakemoney
sudo docker run -p 30001:34321 --name p2 --link=p1:p1 c00w/autoroute:latest -fakemoney -connect p1:34321
```

If you're trying to play with it for testing, I'd reccomend setting up testnet in a box and using
```
autoroute -btc_user=admin1 -btc_pass=123 -btc_host=localhost:19001
```
