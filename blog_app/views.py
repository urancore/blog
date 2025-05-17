from django.shortcuts import get_object_or_404, redirect, render
from django.http import HttpRequest, HttpResponse

from .models import Post, Autor


def main_page(request: HttpRequest):
	posts = Post.objects.all().order_by('-created_at')
	return render(request, "index.html", {'posts': posts})


def read_post(request: HttpRequest, post_id: int):
	post = get_object_or_404(Post, id=post_id)
	data = {"post": post}
	return render(request, "post.html", data)


def create_post(request: HttpRequest):
	if request.method == 'POST':
		title = request.POST.get('post__title')
		content = request.POST.get('post__content')

		if title and content:
			autor = get_object_or_404(Autor, username="test_user")
			post = Post(autor=autor, title=title, content=content)
			post.save()
		return redirect('/')

	return render(request, 'create.html')

def user_page(request: HttpRequest, username: str):
	pass
