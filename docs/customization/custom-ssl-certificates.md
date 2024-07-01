## Custom SSL Certificates

If you need additional domains with custom certificates for your project, you can do it by following this guide.

Let's assume you want to use the `custom-domain.com` domain for your project.

1. Add the custom domain to your `.env` file. You can add multiple domains separated by space.

    ```bash
    TRAEFIK_EXTRA_HOSTS="custom-domain.com"
    ```

2. Copy the certificate and the certificate key to the `~/.reward/ssl/certs` directory:

    ```bash
    cp ~/Downloads/key.pem ~/.reward/ssl/certs/custom-domain.com.key.pem
    cp ~/Downloads/cert.pem ~/.reward/ssl/certs/custom-domain.com.crt.pem
    ```

3. **OPTIONAL**: The hosts from the `TRAEFIK_EXTRA_HOSTS` will be automatically configured and mapped to the webservers.

   However, if you want to finetune the domains (eg: you want to add wildcard domains) create a `.reward/reward-env.yml`
   file with the contents below (this will be additive to the docker compose config Reward uses for the env, anything
   added here will be merged in, and you can see the complete config using `reward env config`):

    ```yaml
    version: "3.5"
    services:
      varnish:
        labels:
          - traefik.http.routers.{{.reward_env_name}}-varnish.rule=
              HostRegexp(`{subdomain:.+}.{{.traefik_domain}}`)
              || Host(`{{.traefik_domain}}`)
              || HostRegexp(`{subdomain:.+}.custom-domain.com`)
              || Host(`custom-domain.com`)
      nginx:
        labels:
          - traefik.http.routers.{{.reward_env_name}}-nginx.rule=
              HostRegexp(`{subdomain:.+}.{{.traefik_domain}}`)
              || Host(`{{.traefik_domain}}`)
              || HostRegexp(`{subdomain:.+}.custom-domain.com`)
              || Host(`custom-domain.com`)
    ```

4. Bring up the environment

    ```bash
    reward env up
    ```

5. Restart Traefik

    ```bash
    reward svc restart traefik
    ```
