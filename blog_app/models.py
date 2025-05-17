from django.db import models

class Autor(models.Model):
	username = models.CharField(max_length=70, unique=True)
	email = models.CharField(max_length=200, unique=True)
	description = models.CharField(max_length=1024, null=True, default="")
	password = models.CharField(max_length=300)

	def __str__(self) -> str:
		return self.username


class Post(models.Model):
	autor = models.ForeignKey(Autor, on_delete=models.CASCADE)
	title = models.CharField(max_length=124, default='none title')
	content = models.CharField(max_length=1024)
	created_at = models.DateTimeField(auto_now_add=True)
	changed_at = models.DateTimeField(auto_now=True)

	def __str__(self) -> str:
		return f"{self.autor} {self.content} {self.created_at}"
