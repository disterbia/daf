
1. rds 생성 후 ec2에서 rds 엔드포인트로 접속해서 db 생성해줘야함.
2. rds 인바운드규칙 설정해도 vpc를 열어줘야 로컬 접속가능. 참조 : https://gksdudrb922.tistory.com/240

https 설정

sudo certbot certonly --standalone -d haruharu-daf.com 

sudo nano /etc/nginx/sites-available/default

server {
    listen 80;
    server_name haruharu-daf.com;

    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl;
    server_name haruharu-daf.com;

    ssl_certificate /etc/letsencrypt/live/haruharu-daf.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/haruharu-daf.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://localhost:40000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 80;
    server_name wellkinson.haruharu-daf.com;

    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl;
    server_name wellkinson.haruharu-daf.com;

    ssl_certificate /etc/letsencrypt/live/haruharu-daf.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/haruharu-daf.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://localhost:50000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

인증서 갱신 : 

sudo systemctl stop nginx
sudo certbot renew --dry-run
sudo nginx -t
sudo systemctl start nginx
sudo certbot renew --webroot -w /var/www/html

 sudo chown -R ubuntu:ubuntu /var/www/flutterweb2
 $ sudo chmod -R 755 /var/www/flutterweb2

 scp -i /Users/admin/Desktop/wellkinson.pem -r web/* ubuntu@43.203.141.35:/var/www/flutterweb2

server {
    listen 80;
    server_name admin.haruharu-daf.com;

    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl;
    server_name admin.haruharu-daf.com;

    ssl_certificate /etc/letsencrypt/live/admin.haruharu-daf.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/admin.haruharu-daf.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers HIGH:!aNULL:!MD5;

    root /var/www/flutterweb;
    index index.html;

    location / {
        try_files $uri /index.html;
    }
}


server {
    listen 80;
    server_name coach.haruharu-daf.com;

    location /.well-known/acme-challenge/ {
        root /var/www/html;
    }

    location / {
        return 301 https://$host$request_uri;
    }
}

server {
    listen 443 ssl;
    server_name coach.haruharu-daf.com;

    ssl_certificate /etc/letsencrypt/live/coach.haruharu-daf.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/coach.haruharu-daf.com/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers HIGH:!aNULL:!MD5;

    root /var/www/flutterweb2;
    index index.html;

    location / {
        try_files $uri /index.html;
    }
}