from django.urls import path
from . import views
urlpatterns = [
	path('', view=views.main_page, name="main_page"),
	path('create/', view=views.create_post, name="create_post"),
	path('edit/<int:post_id>', view=views.edit_post, name="edit_post"),
	path('post/<int:post_id>', view=views.read_post, name="read_post"),
	path('user/<str:username>', view=views.user_page, name="user_page"),
]
