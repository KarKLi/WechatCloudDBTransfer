# 本程序用于将存储在旧utopia小程序云存储的文件的file ID上传到我们自己的服务器，再由服务器批量下载所有文件并上传到新的小程序的云存储上
import requests
import time

#文件夹下的文件file ID
file_list = [
# 'cloud://<your-cloud-database-env-id>/<your-file>',
# 'cloud://<your-cloud-database-env-id>/<your-file>',
# 'cloud://<your-cloud-database-env-id>/<your-file>'
]

# 如果收到的status_code不是200，需要终止进程，并返回text内容查看错误原因
for value in file_list:
    #发起get请求到服务器，path是文件名路径，id是file ID
    r = requests.get('https://<your_upload_URL>?path=file_list/' + str(value[value.rfind('/') + 1 : len(value)]) + '&id=' + str(value))
    print(r.status_code)#获得返回的状态码
    print(r.text)#获得返回的text内容，查看错误信息