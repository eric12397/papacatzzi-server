package db

import "github.com/papacatzzi-server/models"

func (s Store) GetPosts(
	northEastLat float64,
	northEastLng float64,
	southWestLat float64,
	southWestLng float64,
) (posts []models.Post, err error) {

	posts = append(posts,
		models.Post{
			Animal:    "cat",
			PhotoURL:  "",
			Longitude: 33,
			Latitude:  33,
			Author:    "author1",
		},
		models.Post{
			Animal:    "cat",
			PhotoURL:  "",
			Longitude: 33,
			Latitude:  33,
			Author:    "author2",
		},
	)

	return
}
