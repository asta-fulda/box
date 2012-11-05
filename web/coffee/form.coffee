###                                                                                                                                                                                                                 
 * Copyright 2011 Dustin Frisch<fooker@lab.sh>
 * 
 * This file is part of box.
 * 
 * box is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 * 
 * box is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * 
 * You should have received a copy of the GNU General Public License
 * along with box. If not, see <http://www.gnu.org/licenses/>.
###

ko.bindingHandlers.readonly =
    update: (element, valueAccessor) ->
        if ko.utils.unwrapObservable valueAccessor()
            element.setAttribute "readonly", true
        else
            element.removeAttribute "readonly"



class TrackingModel
    constructor: () ->
        @size = ko.observable 0
        @received = ko.observable 0
        
        @state = ko.observable 'starting'
        
        # Generate upload process tracking ID
        @progress_id = (Math.floor(Math.random() * 16).toString(16) for i in [1..32]).reduce (t, s) -> t + s
        
        # Start interval for fetching upload progress
        @interval = setInterval (() =>
                # Fetch the current status of the upload
                $.ajax
                    'url': "/progress?X-Progress-ID=#{@progress_id}"
                    'dataType': 'json'
                    'success': (data) =>
                        # Copy the data to the fields
                        @size data?.size
                        @received data?.received
                        
                        @state data?.state
            ), 1000
        
        
        ko.computed () =>
            # Disable interval if upload has been finished
            if @state() == 'done' or @state() == 'error'
                clearInterval @interval
        
        
        # Compute the progress of the uploading
        @progress = ko.computed () =>
            switch @state()
                when 'starting'
                    '0%'
                when 'done'
                    '100%'
                when 'error'
                    '0%'
                when 'uploading'
                    "#{Math.round(@received() / @size() * 100.0 * 100.0) / 100.0}%"
                else
                    '0%'
        
        
        # Compute the text to display for the tracking state
        @state_text = ko.computed () =>
            switch @state()
                when 'starting'
                    'Warten...'
                when 'done'
                    'Fertig'
                when 'error'
                    'Fehler!'
                when 'uploading'
                    @progress()
                else
                    '...'



class AnswerModel
    constructor: (data) ->
        @id = data.upload_id
        @user = data.upload_user
        @file = data.upload_file
        @size = data.upload_size
        @expiration = data.upload_expiration
        
        @url = "https://box.hs-fulda.org/download/#{@id}?dl=#{@file}"



class ErrorModel
    constructor: (data) ->
        @code = data.code
        @message = data.message



class Model
    constructor: () ->
        @file = ko.observable ''
        @title = ko.observable ''
        @description = ko.observable ''
        @username = ko.observable ''
        @password = ko.observable ''
        @terms_accepted = ko.observable false
        
        # Check if all required fields have some value
        @data_valid = ko.computed () =>
            @file() != '' and
            @username() != '' and
            @password() != '' and
            @terms_accepted()
        
        # The upload tracking object
        @tracking = ko.observable null
        
        # The answer from the upload
        @answer = ko.observable null
        
        # The occurred error - hopefully none
        @error = ko.observable null
        
    
    # Open the file chooser dialog
    open_file_chooser: () =>
        file = document.getElementById "file"
        file.click()
    
    
    # Begin the upload process
    start_upload: (form) =>
        @tracking new TrackingModel
        
        # Get the form data from the form
        data = new FormData form
        
        # Upload the data using an ajax request
        $.ajax
            'url': "/upload?X-Progress-ID=#{@tracking().progress_id}"
            'dataType': 'json'
            'type': 'POST'
            'cache': false
            'processData': false
            'contentType': false
            'data': data
            'username': @username()
            'password': @password()
            'success': (data) =>
                # We got a result from the upload - finish tracking
                @tracking()?.state 'done'
                
                # Store the answer in the model
                @answer new AnswerModel data
                
            'error': (xhr, status, error) =>
                # Stop tracking becaus of the received error
                @tracking()?.state 'error'
                
                @error new ErrorModel
                    'code': xhr.status
                    'message': error
              
        # Suppress the real upload event
        return false
    
    
    # Reset the upload process
    reset_upload: () =>
        # Reset the tracking and the answers
        @tracking null
        @answer null
        @error null
        
        # Sending fake post request to clear authentication cache
        $.ajax
            'url': '/logout'
            'type': 'POST'
            'cache': false
            'processData': false
            'contentType': false
            'username': 'logout'


ko.applyBindings new Model()
