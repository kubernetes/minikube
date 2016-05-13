# Routes
# This file defines all application routes (Higher priority routes first)
# ~~~~

module:testrunner

GET     /                                       App.Index

# Ignore favicon requests
GET     /favicon.ico                            404

# Map static resources from the /app/public folder to the /public path
GET     /public/*filepath                       Static.Serve("public")

# Catch all
*       /:controller/:action                    :controller.:action
